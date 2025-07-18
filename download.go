package main

import (
   "fmt"
   "io"
   "net/http"
   "os"
   "os/exec"
   "path/filepath"
   "regexp"
   "strings"
)

// DownloadTrack downloads a track from YouTube using yt-dlp, converting to the specified format.
// For MP3: Downloads audio and embeds artwork.
// For MP4: Creates a video file with the album artwork as the video stream.
// Returns the path to the downloaded file.
func DownloadTrack(song, artist, artworkURL, format, outDir string, debug bool) (string, error) {
   if _, err := exec.LookPath("yt-dlp"); err != nil {
       return "", fmt.Errorf("yt-dlp not found in PATH: %w", err)
   }
   baseName := sanitizeFileName(fmt.Sprintf("%s - %s", artist, song))
   if err := os.MkdirAll(outDir, 0755); err != nil {
       return "", fmt.Errorf("failed to create output directory: %w", err)
   }
   switch strings.ToLower(format) {
   case "mp3":
       outputTemplate := filepath.Join(outDir, baseName+".%(ext)s")
       args := []string{
           fmt.Sprintf("ytsearch1:%s %s official audio", song, artist),
           "--extract-audio",
           "--audio-format", "mp3",
           "--embed-thumbnail",
           "--add-metadata",
           "--output", outputTemplate,
       }
       cmd := exec.Command("yt-dlp", args...)
       if debug {
           cmd.Stdout = os.Stdout
           cmd.Stderr = os.Stderr
       } else {
           cmd.Stdout = io.Discard
           cmd.Stderr = io.Discard
       }
       if err := cmd.Run(); err != nil {
           return "", err
       }
       return filepath.Join(outDir, baseName+".mp3"), nil
   case "mp4":
       if _, err := exec.LookPath("ffmpeg"); err != nil {
           return "", fmt.Errorf("ffmpeg not found in PATH: %w", err)
       }
       tempDir, err := os.MkdirTemp("", "songdl-*")
       if err != nil {
           return "", fmt.Errorf("failed to create temp dir: %w", err)
       }
       defer os.RemoveAll(tempDir)
       artPath := filepath.Join(tempDir, "cover.jpg")
       if err := downloadFile(artPath, artworkURL); err != nil {
           return "", fmt.Errorf("failed to download artwork: %w", err)
       }
       audioTemplate := filepath.Join(tempDir, "temp_audio.%(ext)s")
       args := []string{
           fmt.Sprintf("ytsearch1:%s %s official audio", song, artist),
           "-f", "bestaudio",
           "--output", audioTemplate,
       }
       cmd := exec.Command("yt-dlp", args...)
       if debug {
           cmd.Stdout = os.Stdout
           cmd.Stderr = os.Stderr
       } else {
           cmd.Stdout = io.Discard
           cmd.Stderr = io.Discard
       }
       if err := cmd.Run(); err != nil {
           return "", fmt.Errorf("audio download failed: %w", err)
       }
       entries, err := os.ReadDir(tempDir)
       if err != nil {
           return "", fmt.Errorf("failed to read temp dir: %w", err)
       }
       var audioFile string
       for _, e := range entries {
           if strings.HasPrefix(e.Name(), "temp_audio") {
               audioFile = filepath.Join(tempDir, e.Name())
               break
           }
       }
       if audioFile == "" {
           return "", fmt.Errorf("audio file not found in temp dir")
       }
       outPath := filepath.Join(outDir, baseName+".mp4")
       ffArgs := []string{
           "-y",
           "-loop", "1",
           "-i", artPath,
           "-i", audioFile,
           "-c:v", "libx264",
           "-tune", "stillimage",
           "-c:a", "aac",
           "-b:a", "192k",
           "-pix_fmt", "yuv420p",
           "-shortest",
           outPath,
       }
       ff := exec.Command("ffmpeg", ffArgs...)
       if debug {
           ff.Stdout = os.Stdout
           ff.Stderr = os.Stderr
       } else {
           ff.Stdout = io.Discard
           ff.Stderr = io.Discard
       }
       if err := ff.Run(); err != nil {
           return "", fmt.Errorf("video creation failed: %w", err)
       }
       return outPath, nil
   default:
       return "", fmt.Errorf("unsupported format: %s", format)
   }
}

// downloadFile downloads a file from the given URL to the specified path.
func downloadFile(path, url string) error {
   resp, err := http.Get(url)
   if err != nil {
       return err
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK {
       return fmt.Errorf("bad status downloading %s: %s", url, resp.Status)
   }
   out, err := os.Create(path)
   if err != nil {
       return err
   }
   defer out.Close()
   _, err = io.Copy(out, resp.Body)
   return err
}

// sanitizeFileName removes characters that are invalid in filenames across different filesystems.
func sanitizeFileName(name string) string {
   invalid := regexp.MustCompile(`[\\/:*?"<>|]`)
   return invalid.ReplaceAllString(name, "_")
}