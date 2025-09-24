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
	hFlag = flag.Bool("h", false, "Show help information")
	helpFlag = flag.Bool("help", false, "Show help information")
)

type Command struct {
	Name        string
	Description string
	Execute     func(args []string) error
}

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
	flag.Parse()

	if *hFlag || *helpFlag {
		printHelp("")
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) > 0 {
		subcommand := args[0]
		
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
		
		if subcommand == "help" && len(args) > 1 {
			printHelp(args[1])
			os.Exit(0)
		}
		
		fmt.Printf("Unknown command: %s\n\n", subcommand)
		printHelp("")
		os.Exit(1)
	}

	err := runDefault()
	if err != nil {
		fmt.Println("An error occurred:", err)
		os.Exit(1)
	}
}

func executeSearch(args []string) error {
   searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
   typeFlag := searchCmd.String("type", "song", "Type of search: song, album, or both (default: song)")
   outFlag := searchCmd.String("out", "downloads", "Output directory for downloaded files")
   debugFlag := searchCmd.Bool("debug", false, "Enable debug logging during download")
   helpFlag := searchCmd.Bool("help", false, "Show help for search command")
   hFlag := searchCmd.Bool("h", false, "Show help for search command")
	
	if err := searchCmd.Parse(args); err != nil {
		return err
	}
	
	if *helpFlag || *hFlag {
		printSearchHelp()
		os.Exit(0)
	}
	
	searchArgs := searchCmd.Args()
	if len(searchArgs) == 0 {
		return fmt.Errorf("search query required")
	}
	
	query := searchArgs[0]
	
	var searchType SearchType
	switch *typeFlag {
	case "song":
		searchType = Song
	case "album":
		searchType = Album
	default:
		searchType = Both
	}
	
   return HandleSearch(query, searchType, *outFlag, *debugFlag)
}

func executeConfig(args []string) error {
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	helpFlag := configCmd.Bool("help", false, "Show help for config command")
	hFlag := configCmd.Bool("h", false, "Show help for config command")
	
	if err := configCmd.Parse(args); err != nil {
		return err
	}
	
	if *helpFlag || *hFlag {
		printConfigHelp()
		os.Exit(0)
	}
	
	fmt.Println("Configuring Apple Music API credentials...")
   return RunOnboarding()
}

func executeDownload(args []string) error {
   downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
   typeFlag := downloadCmd.String("type", "song", "Type of search: song, album, or both (default: song)")
   formatFlag := downloadCmd.String("format", "mp3", "Download format: mp3 or mp4 (default: mp3)")
   outFlag := downloadCmd.String("out", "downloads", "Output directory for downloaded files")
   debugFlag := downloadCmd.Bool("debug", false, "Enable debug logging (show yt-dlp/ffmpeg output)")
   helpFlag := downloadCmd.Bool("help", false, "Show help for download command")
   hFlag := downloadCmd.Bool("h", false, "Show help for download command")

   if err := downloadCmd.Parse(args); err != nil {
       return err
   }
   
   if *helpFlag || *hFlag {
       printDownloadHelp()
       os.Exit(0)
   }

   queryArgs := downloadCmd.Args()
   if len(queryArgs) == 0 {
       return fmt.Errorf("download query required")
   }
   query := strings.Join(queryArgs, " ")

   var searchType SearchType
   switch *typeFlag {
   case "song":
       searchType = Song
   case "album":
       searchType = Album
   default:
       searchType = Song
   }

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

   searcher, err := NewMusicSearcher(config)
   if err != nil {
       return fmt.Errorf("error creating music searcher: %w", err)
   }

   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   results, err := searcher.Search(ctx, query, searchType)
   if err != nil {
       return fmt.Errorf("error searching: %w", err)
   }

   selected, err := DisplaySearchResults(results)
   if err != nil {
       return fmt.Errorf("error selecting result: %w", err)
   }
   fmt.Printf("\nSelected: %s - %s\n", selected.Name, selected.ArtistName)

   fmt.Print("Downloading... ")
   path, err := DownloadTrack(selected.Name, selected.ArtistName, selected.ArtworkURL, *formatFlag, *outFlag, *debugFlag)
   if err != nil {
       return fmt.Errorf("download error: %w", err)
   }
   fmt.Printf("Done. Saved to %s\n", path)
   return nil
}

func executePlaylist(args []string) error {
	playlistCmd := flag.NewFlagSet("playlist", flag.ExitOnError)
	formatFlag := playlistCmd.String("format", "mp3", "Download format: mp3 or mp4 (default: mp3)")
	outFlag := playlistCmd.String("out", "downloads", "Output directory for downloaded files")
	concurrentFlag := playlistCmd.Int("concurrent", 3, "Number of parallel downloads (default: 3)")
	metadataFlag := playlistCmd.Bool("metadata", false, "Save playlist metadata JSON")
	debugFlag := playlistCmd.Bool("debug", false, "Enable debug logging")
	helpFlag := playlistCmd.Bool("help", false, "Show help for playlist command")
	hFlag := playlistCmd.Bool("h", false, "Show help for playlist command")

	if err := playlistCmd.Parse(args); err != nil {
		return err
	}
	
	if *helpFlag || *hFlag {
		printPlaylistHelp()
		os.Exit(0)
	}

	urlArgs := playlistCmd.Args()
	if len(urlArgs) == 0 {
		return fmt.Errorf("Apple Music URL required")
	}
	musicURL := urlArgs[0]

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

	parser := NewPlaylistURLParser()
	resource, err := parser.Parse(musicURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	fmt.Printf("Detected %s from %s storefront\n", resource.Type, resource.Storefront)

	searcher, err := NewExtendedMusicSearcher(config)
	if err != nil {
		return fmt.Errorf("error creating music searcher: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

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

	if err := os.MkdirAll(*outFlag, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if *metadataFlag {
		if err := SavePlaylistMetadata(metadata, *outFlag); err != nil {
			fmt.Printf("Warning: Failed to save metadata: %v\n", err)
		}
	}

	downloader := NewBatchDownloader(*concurrentFlag)
	downloader.Start(ctx)

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

	fmt.Printf("Starting downloads with %d workers...\n\n", *concurrentFlag)
	
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

			if *metadataFlag && metadata != nil {
				metadata.UpdateTrackStatus(result.Job.Track.ID, 
					result.Error == nil, result.FilePath, result.Error)
				SavePlaylistMetadata(metadata, *outFlag)
			}
		}
		done <- true
	}()

	downloader.Close()
	<-done

	downloader.GetProgress().PrintSummary()

	return nil
}

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

func printHelp(command string) {
	switch command {
	case "search":
		printSearchHelp()
	case "download":
		printDownloadHelp()
	case "playlist":
		printPlaylistHelp()
	case "config":
		printConfigHelp()
	default:
		printGeneralHelp()
	}
}

func printGeneralHelp() {
	fmt.Println("Songlink CLI - A powerful tool for music sharing and downloading")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  songlink-cli [flags]                    Process URL from clipboard")
	fmt.Println("  songlink-cli <command> [flags] <args>   Run a specific command")
	fmt.Println("  songlink-cli help <command>             Show help for a command")
	fmt.Println("")
	fmt.Println("COMMANDS:")
	fmt.Println("  search     Search for songs/albums and get shareable links")
	fmt.Println("  download   Search and download tracks as MP3 or MP4 files")
	fmt.Println("  playlist   Download entire playlists or albums from Apple Music")
	fmt.Println("  config     Configure Apple Music API credentials")
	fmt.Println("")
	fmt.Println("GLOBAL FLAGS:")
	fmt.Println("  -h, --help   Show this help message")
	fmt.Println("")
	fmt.Println("URL PROCESSING FLAGS (when run without command):")
	fmt.Println("  -x   Return song.link URL without <> brackets (for Twitter)")
	fmt.Println("  -d   Return song.link URL with <> + Spotify URL (for Discord)")
	fmt.Println("  -s   Return only the Spotify URL")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Process URL from clipboard (default)")
	fmt.Println("  songlink-cli")
	fmt.Println("")
	fmt.Println("  # Get link for Twitter sharing")
	fmt.Println("  songlink-cli -x")
	fmt.Println("")
	fmt.Println("  # Search for a song")
	fmt.Println("  songlink-cli search \"Bohemian Rhapsody\"")
	fmt.Println("")
	fmt.Println("  # Download a track as MP4 with artwork")
	fmt.Println("  songlink-cli download -format=mp4 \"Purple Rain\"")
	fmt.Println("")
	fmt.Println("  # Download an entire album")
	fmt.Println("  songlink-cli playlist \"https://music.apple.com/us/album/...\"")
	fmt.Println("")
	fmt.Println("For more information on a command, run:")
	fmt.Println("  songlink-cli help <command>")
}

func printSearchHelp() {
	fmt.Println("songlink-cli search - Search for songs or albums and get shareable links")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  songlink-cli search [flags] <query>")
	fmt.Println("")
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Search Apple Music for songs or albums and interactively select a result.")
	fmt.Println("  After selection, you can choose to:")
	fmt.Println("    1) Copy shareable links to clipboard")
	fmt.Println("    2) Download the track as MP3")
	fmt.Println("    3) Download as MP4 video with album artwork")
	fmt.Println("")
	fmt.Println("FLAGS:")
	fmt.Println("  -type=<type>   Search type: song, album, or both (default: song)")
	fmt.Println("  -out=<dir>     Output directory for downloads (default: downloads)")
	fmt.Println("  -debug         Enable debug logging during download")
	fmt.Println("")
	fmt.Println("GLOBAL FLAGS (when copying links):")
	fmt.Println("  -x             Format link for Twitter (no brackets)")
	fmt.Println("  -d             Format for Discord (with Spotify URL)")
	fmt.Println("  -s             Copy only Spotify URL")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Search for a song")
	fmt.Println("  songlink-cli search \"Imagine\"")
	fmt.Println("")
	fmt.Println("  # Search for albums only")
	fmt.Println("  songlink-cli search -type=album \"Dark Side of the Moon\"")
	fmt.Println("")
	fmt.Println("  # Search and format for Discord")
	fmt.Println("  songlink-cli search -d \"Hotel California\"")
	fmt.Println("")
	fmt.Println("REQUIREMENTS:")
	fmt.Println("  - Apple Music API credentials (run 'songlink-cli config' to set up)")
	fmt.Println("  - For downloads: yt-dlp and ffmpeg must be installed")
}

func printDownloadHelp() {
	fmt.Println("songlink-cli download - Search and download tracks directly")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  songlink-cli download [flags] <query>")
	fmt.Println("")
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Search for a song or album and download it immediately as an audio")
	fmt.Println("  file (MP3) or video file with album artwork (MP4). This command")
	fmt.Println("  combines search and download into a single step.")
	fmt.Println("")
	fmt.Println("FLAGS:")
	fmt.Println("  -type=<type>     Search type: song or album (default: song)")
	fmt.Println("  -format=<fmt>    Download format: mp3 or mp4 (default: mp3)")
	fmt.Println("  -out=<dir>       Output directory (default: downloads)")
	fmt.Println("  -debug           Show yt-dlp and ffmpeg output")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Download a song as MP3")
	fmt.Println("  songlink-cli download \"Stairway to Heaven\"")
	fmt.Println("")
	fmt.Println("  # Download as MP4 with album artwork")
	fmt.Println("  songlink-cli download -format=mp4 \"Wonderwall\"")
	fmt.Println("")
	fmt.Println("  # Download to custom directory")
	fmt.Println("  songlink-cli download -out=~/Music \"Yesterday\"")
	fmt.Println("")
	fmt.Println("REQUIREMENTS:")
	fmt.Println("  - Apple Music API credentials (run 'songlink-cli config' to set up)")
	fmt.Println("  - yt-dlp: For downloading audio from YouTube")
	fmt.Println("  - ffmpeg: For audio/video processing")
	fmt.Println("")
	fmt.Println("INSTALLATION (macOS):")
	fmt.Println("  brew install yt-dlp ffmpeg")
}

func printPlaylistHelp() {
	fmt.Println("songlink-cli playlist - Download entire playlists or albums")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  songlink-cli playlist [flags] <apple-music-url>")
	fmt.Println("")
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Download all tracks from an Apple Music playlist or album URL.")
	fmt.Println("  Supports parallel downloads and automatic retry on failures.")
	fmt.Println("")
	fmt.Println("SUPPORTED CONTENT:")
	fmt.Println("  ✓ Public catalog albums")
	fmt.Println("  ✓ Public catalog playlists")
	fmt.Println("  ✗ Personal library playlists")
	fmt.Println("  ✗ User-created playlists")
	fmt.Println("  ✗ Radio stations")
	fmt.Println("")
	fmt.Println("FLAGS:")
	fmt.Println("  --format=<fmt>      Download format: mp3 or mp4 (default: mp3)")
	fmt.Println("  --out=<dir>         Output directory (default: downloads)")
	fmt.Println("  --concurrent=<n>    Parallel downloads, 1-10 (default: 3)")
	fmt.Println("  --metadata          Save playlist/album info as JSON")
	fmt.Println("  --debug             Show detailed progress and errors")
	fmt.Println("")
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Download an album")
	fmt.Println("  songlink-cli playlist \"https://music.apple.com/us/album/abbey-road/401469823\"")
	fmt.Println("")
	fmt.Println("  # Download playlist with metadata")
	fmt.Println("  songlink-cli playlist --metadata \"https://music.apple.com/playlist/...\"")
	fmt.Println("")
	fmt.Println("  # Fast download with 5 workers")
	fmt.Println("  songlink-cli playlist --concurrent=5 --format=mp4 \"https://...\"")
	fmt.Println("")
	fmt.Println("FEATURES:")
	fmt.Println("  - Progress tracking for each download")
	fmt.Println("  - Automatic retry with exponential backoff")
	fmt.Println("  - Saves metadata including track status")
	fmt.Println("  - Creates organized directory structure")
	fmt.Println("")
	fmt.Println("TROUBLESHOOTING:")
	fmt.Println("  404 errors: Content may be region-locked, try different")
	fmt.Println("              storefront in URL (e.g., /us/, /gb/, /jp/)")
}

func printConfigHelp() {
	fmt.Println("songlink-cli config - Configure Apple Music API credentials")
	fmt.Println("")
	fmt.Println("USAGE:")
	fmt.Println("  songlink-cli config")
	fmt.Println("")
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Interactive setup wizard for Apple Music API credentials.")
	fmt.Println("  Required for search, download, and playlist features.")
	fmt.Println("")
	fmt.Println("WHAT YOU'LL NEED:")
	fmt.Println("  1. Apple Developer account")
	fmt.Println("  2. MusicKit-enabled API key")
	fmt.Println("  3. Team ID and Key ID")
	fmt.Println("  4. Private key (.p8 file)")
	fmt.Println("")
	fmt.Println("SETUP STEPS:")
	fmt.Println("  1. Sign in to https://developer.apple.com")
	fmt.Println("  2. Go to Certificates, Identifiers & Profiles")
	fmt.Println("  3. Under Keys, create a new key")
	fmt.Println("  4. Enable MusicKit service")
	fmt.Println("  5. Download the .p8 private key file")
	fmt.Println("  6. Note your Team ID and Key ID")
	fmt.Println("")
	fmt.Println("STORED LOCATION:")
	fmt.Println("  ~/.songlink-cli/config.json")
	fmt.Println("")
	fmt.Println("SECURITY:")
	fmt.Println("  Credentials are stored locally and never transmitted")
	fmt.Println("  except to Apple's API servers.")
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
