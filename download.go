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
   "time"
)

func DownloadTrack(song, artist, artworkURL, format, outDir string, debug bool) (string, error) {
   ytdlpPath, err := exec.LookPath("yt-dlp")
   if err != nil {
       return "", fmt.Errorf("yt-dlp not found in PATH. Please install it: brew install yt-dlp (macOS) or see README for other systems")
   }

   if err := checkYtDlpVersion(ytdlpPath, debug); err != nil {
       return "", err
   }
   baseName := sanitizeFileName(fmt.Sprintf("%s - %s", artist, song))
   if err := os.MkdirAll(outDir, 0755); err != nil {
       return "", fmt.Errorf("failed to create output directory: %w", err)
   }
   switch strings.ToLower(format) {
   case "mp3":
       outputTemplate := filepath.Join(outDir, baseName+".%(ext)s")
       searchQueries := []string{
           fmt.Sprintf("ytsearch1:%s %s lyrics", song, artist),
           fmt.Sprintf("ytsearch1:%s %s topic", song, artist),
           fmt.Sprintf("ytsearch1:%s %s", song, artist),
       }

       var lastErr error
       for _, query := range searchQueries {
           args := []string{
               query,
               "--extract-audio",
               "--audio-format", "mp3",
               "--audio-quality", "192K",
               "--embed-thumbnail",
               "--add-metadata",
               "--output", outputTemplate,
               "--no-check-certificates",
               "--no-playlist",
               "--no-warnings",
               "--ignore-errors",
           }
           cmd := exec.Command("yt-dlp", args...)
           if debug {
               fmt.Printf("Trying search: %s\n", query)
               cmd.Stdout = os.Stdout
               cmd.Stderr = os.Stderr
           } else {
               cmd.Stdout = io.Discard
               cmd.Stderr = io.Discard
           }
           err := cmd.Run()
           if err == nil {
               return filepath.Join(outDir, baseName+".mp3"), nil
           }
           lastErr = err
       }
       if strings.Contains(fmt.Sprint(lastErr), "Requested format is not available") ||
          strings.Contains(fmt.Sprint(lastErr), "Signature extraction failed") {
           return "", fmt.Errorf("download failed - yt-dlp is likely outdated. Please update: yt-dlp -U or brew upgrade yt-dlp")
       }
       return "", fmt.Errorf("all download attempts failed (try --debug for details): %w", lastErr)
   case "mp4":
       if _, err := exec.LookPath("ffmpeg"); err != nil {
           return "", fmt.Errorf("ffmpeg not found in PATH. Please install it: brew install ffmpeg (macOS) or see README")
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
       searchQueries := []string{
           fmt.Sprintf("ytsearch1:%s %s lyrics", song, artist),
           fmt.Sprintf("ytsearch1:%s %s topic", song, artist),
           fmt.Sprintf("ytsearch1:%s %s", song, artist),
       }

       var lastErr error
       for _, query := range searchQueries {
           args := []string{
               query,
               "--extract-audio",
               "--audio-format", "m4a",
               "--audio-quality", "192K",
               "--output", audioTemplate,
               "--no-check-certificates",
               "--no-playlist",
               "--no-warnings",
               "--ignore-errors",
           }
           cmd := exec.Command("yt-dlp", args...)
           if debug {
               fmt.Printf("Trying search: %s\n", query)
               cmd.Stdout = os.Stdout
               cmd.Stderr = os.Stderr
           } else {
               cmd.Stdout = io.Discard
               cmd.Stderr = io.Discard
           }
           err := cmd.Run()
           if err == nil {
               break
           }
           lastErr = err
       }
       if lastErr != nil {
           if strings.Contains(fmt.Sprint(lastErr), "Requested format is not available") ||
              strings.Contains(fmt.Sprint(lastErr), "Signature extraction failed") {
               return "", fmt.Errorf("download failed - yt-dlp is likely outdated. Please update: yt-dlp -U or brew upgrade yt-dlp")
           }
           return "", fmt.Errorf("audio download failed (try --debug for details): %w", lastErr)
       }
       entries, err := os.ReadDir(tempDir)
       if err != nil {
           return "", fmt.Errorf("failed to read temp dir: %w", err)
       }
       var audioFile string
       for _, e := range entries {
           if strings.HasPrefix(e.Name(), "temp_audio") && !strings.HasSuffix(e.Name(), ".jpg") && !strings.HasSuffix(e.Name(), ".png") && !strings.HasSuffix(e.Name(), ".webp") {
               audioFile = filepath.Join(tempDir, e.Name())
               break
           }
       }
       if audioFile == "" {
           return "", fmt.Errorf("audio extraction failed - no audio file was created. This may indicate yt-dlp needs updating")
       }
       outPath := filepath.Join(outDir, baseName+".mp4")
       ffArgs := []string{
           "-y",
           "-loop", "1",
           "-framerate", "1",
           "-i", artPath,
           "-i", audioFile,
           "-c:v", "libx264",
           "-preset", "medium",
           "-tune", "stillimage",
           "-c:a", "aac",
           "-b:a", "192k",
           "-pix_fmt", "yuv420p",
           "-shortest",
           "-movflags", "+faststart",
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
           return "", fmt.Errorf("video creation failed with ffmpeg. Ensure ffmpeg is properly installed: %w", err)
       }
       return outPath, nil
   default:
       return "", fmt.Errorf("unsupported format: %s", format)
   }
}

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

func sanitizeFileName(name string) string {
   invalid := regexp.MustCompile(`[\\/:*?"<>|]`)
   return invalid.ReplaceAllString(name, "_")
}

func checkYtDlpVersion(ytdlpPath string, debug bool) error {
   cmd := exec.Command(ytdlpPath, "--version")
   output, err := cmd.Output()
   if err != nil {
       return fmt.Errorf("failed to check yt-dlp version: %w", err)
   }

   version := strings.TrimSpace(string(output))
   if debug {
       fmt.Printf("yt-dlp version: %s\n", version)
   }

   parts := strings.Split(version, ".")
   if len(parts) >= 3 {
       yearStr := parts[0]
       monthStr := parts[1]

       year := 0
       month := 0
       fmt.Sscanf(yearStr, "%d", &year)
       fmt.Sscanf(monthStr, "%d", &month)

       now := time.Now()
       currentYear := now.Year()
       currentMonth := int(now.Month())

       monthsOld := (currentYear - year) * 12 + (currentMonth - month)
       if monthsOld > 2 {
           return fmt.Errorf("yt-dlp version %s is too old (>2 months). YouTube frequently breaks compatibility. Please update: yt-dlp -U or brew upgrade yt-dlp", version)
       }
       if monthsOld > 1 && debug {
           fmt.Printf("Warning: yt-dlp version %s is %d month(s) old. Consider updating for best compatibility.\n", version, monthsOld)
       }
   }

   return nil
}