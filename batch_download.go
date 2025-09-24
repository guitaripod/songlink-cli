package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type BatchDownloader struct {
	concurrency   int
	downloadQueue chan DownloadJob
	results       chan DownloadResult
	wg            sync.WaitGroup
	progress      *ProgressTracker
}

type DownloadJob struct {
	Track      SearchResult
	Format     string
	OutputDir  string
	Debug      bool
	RetryCount int
	Index      int
}

type DownloadResult struct {
	Job      DownloadJob
	FilePath string
	Error    error
	Duration time.Duration
}

func NewBatchDownloader(concurrency int) *BatchDownloader {
	if concurrency <= 0 {
		concurrency = 3
	}
	
	return &BatchDownloader{
		concurrency:   concurrency,
		downloadQueue: make(chan DownloadJob, 100),
		results:       make(chan DownloadResult, 100),
		progress:      NewProgressTracker(),
	}
}

func (bd *BatchDownloader) Start(ctx context.Context) {
	for i := 0; i < bd.concurrency; i++ {
		bd.wg.Add(1)
		go bd.worker(ctx, i)
	}
}

func (bd *BatchDownloader) worker(ctx context.Context, workerID int) {
	defer bd.wg.Done()
	
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-bd.downloadQueue:
			if !ok {
				return
			}
			
			bd.progress.StartDownload(job.Track.ID, job.Track.Name, job.Track.ArtistName)
			
			startTime := time.Now()
			filePath, err := bd.downloadTrack(job)
			duration := time.Since(startTime)
			
			result := DownloadResult{
				Job:      job,
				FilePath: filePath,
				Error:    err,
				Duration: duration,
			}
			
			if err != nil {
				bd.progress.MarkFailed(job.Track.ID, err)
			} else {
				bd.progress.MarkCompleted(job.Track.ID, filePath)
			}
			
			select {
			case bd.results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (bd *BatchDownloader) downloadTrack(job DownloadJob) (string, error) {
	maxRetries := 3
	retryDelay := time.Second
	
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
		}
		
		filePath, err := DownloadTrack(
			job.Track.Name,
			job.Track.ArtistName,
			job.Track.ArtworkURL,
			job.Format,
			job.OutputDir,
			job.Debug,
		)
		
		if err == nil {
			return filePath, nil
		}
		
		lastErr = err
		bd.progress.UpdateRetry(job.Track.ID, attempt+1, maxRetries)
	}
	
	return "", fmt.Errorf("download failed after %d attempts: %w", maxRetries+1, lastErr)
}

func (bd *BatchDownloader) QueueDownload(job DownloadJob) error {
	select {
	case bd.downloadQueue <- job:
		bd.progress.QueueTrack(job.Track.ID, job.Track.Name, job.Track.ArtistName)
		return nil
	default:
		return fmt.Errorf("download queue is full")
	}
}

func (bd *BatchDownloader) Close() {
	close(bd.downloadQueue)
	bd.wg.Wait()
	close(bd.results)
}

func (bd *BatchDownloader) GetResults() <-chan DownloadResult {
	return bd.results
}

func (bd *BatchDownloader) GetProgress() *ProgressTracker {
	return bd.progress
}

type ProgressTracker struct {
	mu         sync.RWMutex
	total      int32
	completed  int32
	failed     int32
	inProgress map[string]*TrackProgress
	startTime  time.Time
}

type TrackProgress struct {
	ID         string
	Name       string
	Artist     string
	Status     DownloadStatus
	FilePath   string
	Error      error
	StartTime  time.Time
	EndTime    time.Time
	RetryCount int
}

type DownloadStatus string

const (
	StatusQueued      DownloadStatus = "queued"
	StatusDownloading DownloadStatus = "downloading"
	StatusCompleted   DownloadStatus = "completed"
	StatusFailed      DownloadStatus = "failed"
	StatusRetrying    DownloadStatus = "retrying"
)

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		inProgress: make(map[string]*TrackProgress),
		startTime:  time.Now(),
	}
}

func (pt *ProgressTracker) QueueTrack(id, name, artist string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	atomic.AddInt32(&pt.total, 1)
	pt.inProgress[id] = &TrackProgress{
		ID:     id,
		Name:   name,
		Artist: artist,
		Status: StatusQueued,
	}
}

func (pt *ProgressTracker) StartDownload(id, name, artist string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if track, ok := pt.inProgress[id]; ok {
		track.Status = StatusDownloading
		track.StartTime = time.Now()
	}
}

func (pt *ProgressTracker) UpdateRetry(id string, attempt, maxAttempts int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if track, ok := pt.inProgress[id]; ok {
		track.Status = StatusRetrying
		track.RetryCount = attempt
	}
}

func (pt *ProgressTracker) MarkCompleted(id, filePath string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	atomic.AddInt32(&pt.completed, 1)
	if track, ok := pt.inProgress[id]; ok {
		track.Status = StatusCompleted
		track.FilePath = filePath
		track.EndTime = time.Now()
	}
}

func (pt *ProgressTracker) MarkFailed(id string, err error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	atomic.AddInt32(&pt.failed, 1)
	if track, ok := pt.inProgress[id]; ok {
		track.Status = StatusFailed
		track.Error = err
		track.EndTime = time.Now()
	}
}

func (pt *ProgressTracker) GetStats() (total, completed, failed int32) {
	return atomic.LoadInt32(&pt.total),
		atomic.LoadInt32(&pt.completed),
		atomic.LoadInt32(&pt.failed)
}

func (pt *ProgressTracker) GetTrackProgress(id string) (*TrackProgress, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	track, ok := pt.inProgress[id]
	return track, ok
}

func (pt *ProgressTracker) GetAllProgress() map[string]*TrackProgress {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	result := make(map[string]*TrackProgress)
	for k, v := range pt.inProgress {
		result[k] = v
	}
	return result
}

func (pt *ProgressTracker) PrintSummary() {
	total, completed, failed := pt.GetStats()
	duration := time.Since(pt.startTime)
	
	fmt.Printf("\n========== Download Summary ==========\n")
	fmt.Printf("Total tracks: %d\n", total)
	fmt.Printf("Completed: %d\n", completed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Duration: %s\n", duration.Round(time.Second))
	fmt.Printf("=====================================\n")
	
	if failed > 0 {
		fmt.Printf("\nFailed downloads:\n")
		for _, track := range pt.GetAllProgress() {
			if track.Status == StatusFailed {
				fmt.Printf("- %s - %s: %v\n", track.Artist, track.Name, track.Error)
			}
		}
	}
}