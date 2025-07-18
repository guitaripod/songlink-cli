package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PlaylistMetadata represents metadata for a downloaded playlist/album
type PlaylistMetadata struct {
	Type        string           `json:"type"` // "album" or "playlist"
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Artist      string           `json:"artist,omitempty"`      // For albums
	Curator     string           `json:"curator,omitempty"`     // For playlists
	Description string           `json:"description,omitempty"`
	TrackCount  int              `json:"track_count"`
	Duration    int              `json:"duration_seconds"`
	ArtworkURL  string           `json:"artwork_url"`
	SourceURL   string           `json:"source_url"`
	DownloadedAt time.Time       `json:"downloaded_at"`
	Tracks      []TrackMetadata  `json:"tracks"`
}

// TrackMetadata represents metadata for a single track
type TrackMetadata struct {
	Index       int    `json:"index"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Artist      string `json:"artist"`
	Duration    int    `json:"duration_seconds,omitempty"`
	FilePath    string `json:"file_path"`
	Downloaded  bool   `json:"downloaded"`
	DownloadedAt *time.Time `json:"downloaded_at,omitempty"`
	Error       string `json:"error,omitempty"`
}

// SavePlaylistMetadata saves playlist metadata to a JSON file
func SavePlaylistMetadata(metadata *PlaylistMetadata, outputDir string) error {
	// Create filename based on playlist/album name
	filename := sanitizeFileName(metadata.Name) + "_metadata.json"
	filepath := filepath.Join(outputDir, filename)
	
	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	
	return nil
}

// LoadPlaylistMetadata loads playlist metadata from a JSON file
func LoadPlaylistMetadata(filepath string) (*PlaylistMetadata, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}
	
	var metadata PlaylistMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	return &metadata, nil
}

// UpdateTrackStatus updates the download status of a track in metadata
func (pm *PlaylistMetadata) UpdateTrackStatus(trackID string, downloaded bool, filepath string, err error) {
	for i, track := range pm.Tracks {
		if track.ID == trackID {
			pm.Tracks[i].Downloaded = downloaded
			if downloaded {
				now := time.Now()
				pm.Tracks[i].DownloadedAt = &now
				pm.Tracks[i].FilePath = filepath
			}
			if err != nil {
				pm.Tracks[i].Error = err.Error()
			}
			break
		}
	}
}

// CreateAlbumMetadata creates metadata for an album
func CreateAlbumMetadata(album *AlbumWithTracks, sourceURL string) *PlaylistMetadata {
	metadata := &PlaylistMetadata{
		Type:         "album",
		ID:           album.ID,
		Name:         album.Name,
		Artist:       album.ArtistName,
		TrackCount:   album.TrackCount,
		ArtworkURL:   album.ArtworkURL,
		SourceURL:    sourceURL,
		DownloadedAt: time.Now(),
		Tracks:       make([]TrackMetadata, len(album.Tracks)),
	}
	
	// Add track metadata
	for i, track := range album.Tracks {
		metadata.Tracks[i] = TrackMetadata{
			Index:      i + 1,
			ID:         track.ID,
			Name:       track.Name,
			Artist:     track.ArtistName,
			Downloaded: false,
		}
	}
	
	return metadata
}

// CreatePlaylistMetadata creates metadata for a playlist
func CreatePlaylistMetadata(playlist *PlaylistWithTracks, sourceURL string) *PlaylistMetadata {
	metadata := &PlaylistMetadata{
		Type:         "playlist",
		ID:           playlist.ID,
		Name:         playlist.Name,
		Curator:      playlist.CuratorName,
		TrackCount:   playlist.TrackCount,
		ArtworkURL:   playlist.ArtworkURL,
		SourceURL:    sourceURL,
		DownloadedAt: time.Now(),
		Tracks:       make([]TrackMetadata, len(playlist.Tracks)),
	}
	
	// Add track metadata
	for i, track := range playlist.Tracks {
		metadata.Tracks[i] = TrackMetadata{
			Index:      i + 1,
			ID:         track.ID,
			Name:       track.Name,
			Artist:     track.ArtistName,
			Downloaded: false,
		}
	}
	
	return metadata
}