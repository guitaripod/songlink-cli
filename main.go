package main

import (
   "context"
   "flag"
   "fmt"
   "os"
   "strings"
   "sync"
   "time"

   "github.com/atotto/clipboard"
)

var (
	xFlag = flag.Bool("x", false, "Return the song.link URL without surrounding <>")
	dFlag = flag.Bool("d", false, "Return the song.link URL surrounded by <> and the Spotify URL")
	sFlag = flag.Bool("s", false, "Return only the Spotify URL")
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Execute     func(args []string) error
}

// Commands available in the application
var commands = []Command{
	{
		Name:        "search",
		Description: "Search for a song or album and get its links",
		Execute:     executeSearch,
	},
   {
       Name:        "config",
       Description: "Configure Apple Music API credentials",
       Execute:     executeConfig,
   },
   {
       Name:        "download",
       Description: "Search for a song or album and download it as mp3 or mp4",
       Execute:     executeDownload,
   },
   {
       Name:        "playlist",
       Description: "Download an entire playlist or album from Apple Music URL",
       Execute:     executePlaylist,
   },
}

func main() {
	// Define base flags
	flag.Parse()

	// Check if a subcommand is provided
	args := flag.Args()
	if len(args) > 0 {
		subcommand := args[0]
		
		// Find and execute the appropriate command
		for _, cmd := range commands {
			if cmd.Name == subcommand {
				err := cmd.Execute(args[1:])
				if err != nil {
					fmt.Println("An error occurred:", err)
					os.Exit(1)
				}
				return
			}
		}
		
		// If we get here, the subcommand wasn't recognized
		fmt.Printf("Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	// No subcommand provided, run the default behavior
	err := runDefault()
	if err != nil {
		fmt.Println("An error occurred:", err)
		os.Exit(1)
	}
}

// executeSearch handles the search subcommand
func executeSearch(args []string) error {
   // Define search flags
   searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
   typeFlag := searchCmd.String("type", "song", "Type of search: song, album, or both (default: song)")
   outFlag := searchCmd.String("out", "downloads", "Output directory for downloaded files")
   debugFlag := searchCmd.Bool("debug", false, "Enable debug logging during download")
	
	// Parse search flags
	if err := searchCmd.Parse(args); err != nil {
		return err
	}
	
	// Get search query
	searchArgs := searchCmd.Args()
	if len(searchArgs) == 0 {
		return fmt.Errorf("search query required")
	}
	
	query := searchArgs[0]
	
	// Determine search type
	var searchType SearchType
	switch *typeFlag {
	case "song":
		searchType = Song
	case "album":
		searchType = Album
	default:
		// Use Both to search for songs and albums
		searchType = Both
	}
	
   // Handle search
   return HandleSearch(query, searchType, *outFlag, *debugFlag)
}

// executeConfig handles the config subcommand
func executeConfig(args []string) error {
	fmt.Println("Configuring Apple Music API credentials...")
   return RunOnboarding()
}

// executeDownload handles the download subcommand
func executeDownload(args []string) error {
   // Define download flags
   downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
   typeFlag := downloadCmd.String("type", "song", "Type of search: song, album, or both (default: song)")
   formatFlag := downloadCmd.String("format", "mp3", "Download format: mp3 or mp4 (default: mp3)")
   outFlag := downloadCmd.String("out", "downloads", "Output directory for downloaded files")
   debugFlag := downloadCmd.Bool("debug", false, "Enable debug logging (show yt-dlp/ffmpeg output)")

   // Parse flags
   if err := downloadCmd.Parse(args); err != nil {
       return err
   }

   // Get search query
   queryArgs := downloadCmd.Args()
   if len(queryArgs) == 0 {
       return fmt.Errorf("download query required")
   }
   query := strings.Join(queryArgs, " ")

   // Determine search type
   var searchType SearchType
   switch *typeFlag {
   case "song":
       searchType = Song
   case "album":
       searchType = Album
   default:
       searchType = Song
   }

   // Load config
   config, err := LoadConfig()
   if err != nil {
       return fmt.Errorf("error loading config: %w", err)
   }
   if !config.ConfigExists {
       fmt.Println("Apple Music API credentials not found. Let's set them up.")
       if err := RunOnboarding(); err != nil {
           return fmt.Errorf("error during onboarding: %w", err)
       }
       config, err = LoadConfig()
       if err != nil {
           return fmt.Errorf("error loading config after onboarding: %w", err)
       }
   }

   // Create music searcher
   searcher, err := NewMusicSearcher(config)
   if err != nil {
       return fmt.Errorf("error creating music searcher: %w", err)
   }

   // Search for music
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   results, err := searcher.Search(ctx, query, searchType)
   if err != nil {
       return fmt.Errorf("error searching: %w", err)
   }

   // Display results and select
   selected, err := DisplaySearchResults(results)
   if err != nil {
       return fmt.Errorf("error selecting result: %w", err)
   }
   fmt.Printf("\nSelected: %s - %s\n", selected.Name, selected.ArtistName)

   // Download track via YouTube
   fmt.Print("Downloading... ")
   path, err := DownloadTrack(selected.Name, selected.ArtistName, selected.ArtworkURL, *formatFlag, *outFlag, *debugFlag)
   if err != nil {
       return fmt.Errorf("download error: %w", err)
   }
   fmt.Printf("Done. Saved to %s\n", path)
   return nil
}

// executePlaylist handles the playlist subcommand
func executePlaylist(args []string) error {
	// Define playlist flags
	playlistCmd := flag.NewFlagSet("playlist", flag.ExitOnError)
	formatFlag := playlistCmd.String("format", "mp3", "Download format: mp3 or mp4 (default: mp3)")
	outFlag := playlistCmd.String("out", "downloads", "Output directory for downloaded files")
	concurrentFlag := playlistCmd.Int("concurrent", 3, "Number of parallel downloads (default: 3)")
	metadataFlag := playlistCmd.Bool("metadata", false, "Save playlist metadata JSON")
	debugFlag := playlistCmd.Bool("debug", false, "Enable debug logging")

	// Parse flags
	if err := playlistCmd.Parse(args); err != nil {
		return err
	}

	// Get URL argument
	urlArgs := playlistCmd.Args()
	if len(urlArgs) == 0 {
		return fmt.Errorf("Apple Music URL required")
	}
	musicURL := urlArgs[0]

	// Load config
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}
	if !config.ConfigExists {
		fmt.Println("Apple Music API credentials not found. Let's set them up.")
		if err := RunOnboarding(); err != nil {
			return fmt.Errorf("error during onboarding: %w", err)
		}
		config, err = LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config after onboarding: %w", err)
		}
	}

	// Parse the URL
	parser := NewPlaylistURLParser()
	resource, err := parser.Parse(musicURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	fmt.Printf("Detected %s from %s storefront\n", resource.Type, resource.Storefront)

	// Create extended music searcher
	searcher, err := NewExtendedMusicSearcher(config)
	if err != nil {
		return fmt.Errorf("error creating music searcher: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Fetch tracks based on resource type
	var tracks []SearchResult
	var metadata *PlaylistMetadata

	switch resource.Type {
	case ParsedAlbum:
		fmt.Printf("Fetching album details...\n")
		album, err := searcher.GetAlbumWithTracks(ctx, resource.ID, resource.Storefront)
		if err != nil {
			return fmt.Errorf("error fetching album: %w", err)
		}
		tracks = album.Tracks
		metadata = CreateAlbumMetadata(album, musicURL)
		fmt.Printf("Album: %s - %s (%d tracks)\n", album.Name, album.ArtistName, len(tracks))

	case ParsedPlaylist:
		fmt.Printf("Fetching playlist details...\n")
		playlist, err := searcher.GetPlaylistWithTracks(ctx, resource.ID, resource.Storefront)
		if err != nil {
			return fmt.Errorf("error fetching playlist: %w", err)
		}
		tracks = playlist.Tracks
		metadata = CreatePlaylistMetadata(playlist, musicURL)
		fmt.Printf("Playlist: %s by %s (%d tracks)\n", playlist.Name, playlist.CuratorName, len(tracks))
	}

	if len(tracks) == 0 {
		return fmt.Errorf("no tracks found")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outFlag, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save initial metadata if requested
	if *metadataFlag {
		if err := SavePlaylistMetadata(metadata, *outFlag); err != nil {
			fmt.Printf("Warning: Failed to save metadata: %v\n", err)
		}
	}

	// Create batch downloader
	downloader := NewBatchDownloader(*concurrentFlag)
	downloader.Start(ctx)

	// Queue all downloads
	fmt.Printf("\nQueuing %d tracks for download...\n", len(tracks))
	for i, track := range tracks {
		job := DownloadJob{
			Track:     track,
			Format:    *formatFlag,
			OutputDir: *outFlag,
			Debug:     *debugFlag,
			Index:     i + 1,
		}
		if err := downloader.QueueDownload(job); err != nil {
			fmt.Printf("Failed to queue track %d: %v\n", i+1, err)
		}
	}

	// Monitor downloads
	fmt.Printf("Starting downloads with %d workers...\n\n", *concurrentFlag)
	
	// Create a channel to signal when all downloads are processed
	done := make(chan bool)
	go func() {
		for result := range downloader.GetResults() {
			if result.Error != nil {
				fmt.Printf("❌ [%d/%d] Failed: %s - %s (%v)\n", 
					result.Job.Index, len(tracks),
					result.Job.Track.ArtistName, result.Job.Track.Name, 
					result.Error)
			} else {
				fmt.Printf("✅ [%d/%d] Downloaded: %s - %s\n", 
					result.Job.Index, len(tracks),
					result.Job.Track.ArtistName, result.Job.Track.Name)
			}

			// Update metadata if enabled
			if *metadataFlag && metadata != nil {
				metadata.UpdateTrackStatus(result.Job.Track.ID, 
					result.Error == nil, result.FilePath, result.Error)
				// Save updated metadata
				SavePlaylistMetadata(metadata, *outFlag)
			}
		}
		done <- true
	}()

	// Close the downloader and wait for completion
	downloader.Close()
	<-done

	// Print summary
	downloader.GetProgress().PrintSummary()

	return nil
}

// runDefault runs the default behavior (process URL from clipboard)
func runDefault() error {
	searchURL, err := clipboard.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading clipboard: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	stopLoading := make(chan bool)
	go func() {
		defer wg.Done()
		loadingIndicator(stopLoading)
	}()

	err = GetLinks(searchURL)
	if err != nil {
		return fmt.Errorf("error getting links: %w", err)
	}

	stopLoading <- true
	wg.Wait()

	return nil
}

// printUsage prints usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  songlink-cli [flags]                 Process URL from clipboard")
	fmt.Println("  songlink-cli search [flags] <query>  Search for a song or album")
	fmt.Println("  songlink-cli config                  Configure Apple Music API credentials")
	fmt.Println("  songlink-cli download [flags] <query> Download a song or album")
	fmt.Println("  songlink-cli playlist [flags] <url>  Download entire playlist/album")
	fmt.Println("\nFlags:")
	fmt.Println("  -x  Return the song.link URL without surrounding <>")
	fmt.Println("  -d  Return the song.link URL surrounded by <> and the Spotify URL")
	fmt.Println("  -s  Return only the Spotify URL")
	fmt.Println("\nSearch Flags:")
	fmt.Println("  -type=<type>  Type of search: song, album, or both (default: song)")
	fmt.Println("\nPlaylist Flags:")
	fmt.Println("  --format=<fmt>    Download format: mp3 or mp4 (default: mp3)")
	fmt.Println("  --out=<dir>       Output directory (default: downloads)")
	fmt.Println("  --concurrent=<n>  Number of parallel downloads (default: 3)")
	fmt.Println("  --metadata        Save playlist metadata JSON")
	fmt.Println("  --debug           Show debug output")
}

func loadingIndicator(stop chan bool) {
	chars := []string{"-", "\\", "|", "/"}
	i := 0
	for {
		select {
		case <-stop:
			fmt.Print("\r")
			return
		default:
			fmt.Printf("\rLoading %s", chars[i])
			i = (i + 1) % len(chars)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
