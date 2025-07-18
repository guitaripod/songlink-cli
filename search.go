package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/guitaripod/musickitkat"
	"github.com/guitaripod/musickitkat/auth"
)

// SearchType represents the type of content to search for.
type SearchType string

const (
	Song  SearchType = "song"
	Album SearchType = "album"
	Both  SearchType = "both"
)

// MusicSearcher provides methods for searching Apple Music content.
type MusicSearcher struct {
	client *musickitkat.Client
}

// SearchResult represents a single search result from Apple Music.
type SearchResult struct {
	ID         string
	Name       string
	ArtistName string
	Type       SearchType
	URL        string
	ArtworkURL string
}

func NewMusicSearcher(config *Config) (*MusicSearcher, error) {
	if !config.ConfigExists {
		return nil, errors.New("apple music api credentials not configured")
	}

	developerToken, err := auth.NewDeveloperToken(
		config.TeamID,
		config.KeyID,
		[]byte(config.PrivateKey),
		config.MusicID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create developer token: %w", err)
	}

	client := musickitkat.NewClient(
		musickitkat.WithDeveloperToken(developerToken),
	)

	return &MusicSearcher{
		client: client,
	}, nil
}

func (ms *MusicSearcher) Search(ctx context.Context, query string, searchType SearchType) ([]SearchResult, error) {
	var results []SearchResult
	var searchTypes []string

	switch searchType {
	case Song:
		searchTypes = []string{string(musickitkat.SearchTypesSongs)}
	case Album:
		searchTypes = []string{string(musickitkat.SearchTypesAlbums)}
	case Both:
		searchTypes = []string{string(musickitkat.SearchTypesSongs), string(musickitkat.SearchTypesAlbums)}
	default:
		searchTypes = []string{string(musickitkat.SearchTypesSongs)}
	}

	for _, st := range searchTypes {
		searchResults, err := ms.client.Search.Search(ctx, query, []string{st}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to search %s: %w", st, err)
		}

		if st == string(musickitkat.SearchTypesSongs) && len(searchResults.Results.Songs.Data) > 0 {
			for _, song := range searchResults.Results.Songs.Data {
				artURL := song.Attributes.Artwork.URL
				artURL = strings.ReplaceAll(artURL, "{w}", "500")
				artURL = strings.ReplaceAll(artURL, "{h}", "500")
				results = append(results, SearchResult{
					ID:         song.ID,
					Name:       song.Attributes.Name,
					ArtistName: song.Attributes.ArtistName,
					Type:       Song,
					URL:        song.Attributes.URL,
					ArtworkURL: artURL,
				})
			}
		}

		if st == string(musickitkat.SearchTypesAlbums) && len(searchResults.Results.Albums.Data) > 0 {
			for _, album := range searchResults.Results.Albums.Data {
				artURL := album.Attributes.Artwork.URL
				artURL = strings.ReplaceAll(artURL, "{w}", "500")
				artURL = strings.ReplaceAll(artURL, "{h}", "500")
				results = append(results, SearchResult{
					ID:         album.ID,
					Name:       album.Attributes.Name,
					ArtistName: album.Attributes.ArtistName,
					Type:       Album,
					URL:        album.Attributes.URL,
					ArtworkURL: artURL,
				})
			}
		}
	}

	return results, nil
}

func DisplaySearchResults(results []SearchResult) (*SearchResult, error) {
	if len(results) == 0 {
		return nil, errors.New("no results found")
	}

	fmt.Println("\nSearch Results:")
	fmt.Println("----------------")

	for i, result := range results {
		typeStr := "Song"
		if result.Type == Album {
			typeStr = "Album"
		}
		fmt.Printf("%d. [%s] %s - %s\n", i+1, typeStr, result.Name, result.ArtistName)
	}

	var choice int
	fmt.Print("\nSelect a result (1-", len(results), "): ")
	
	var input string
	fmt.Scanln(&input)
	
	if input == "" {
		fmt.Println("1 (automatic selection)")
		choice = 1
	} else {
		_, err := fmt.Sscanf(input, "%d", &choice)
		if err != nil || choice < 1 || choice > len(results) {
			return nil, errors.New("invalid selection")
		}
	}

	return &results[choice-1], nil
}

func HandleSearch(query string, searchType SearchType, outDir string, debug bool) error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	if !config.ConfigExists {
		fmt.Println("Apple Music API credentials not found. Let's set them up.")
		err = RunOnboarding()
		if err != nil {
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

	stopLoading := make(chan bool)
	go func() {
		loadingIndicator(stopLoading)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := searcher.Search(ctx, query, searchType)

	stopLoading <- true

	if err != nil {
		return fmt.Errorf("error searching: %w", err)
	}

   selected, err := DisplaySearchResults(results)
	if err != nil {
		return fmt.Errorf("error selecting result: %w", err)
	}

   fmt.Printf("\nSelected: %s - %s\n", selected.Name, selected.ArtistName)
   fmt.Println("\nWhat would you like to do?")
   fmt.Println("1) Copy song.link + Spotify URL to clipboard")
   fmt.Println("2) Download MP3")
   fmt.Println("3) Download MP4 (video with artwork)")
   fmt.Print("Enter choice (1-3, default 1): ")
   var choice string
   fmt.Scanln(&choice)
   switch choice {
   case "", "1":
       if err := GetLinks(selected.URL); err != nil {
           return fmt.Errorf("error getting links: %w", err)
       }
   case "2":
       fmt.Print("Downloading MP3... ")
       path, err := DownloadTrack(selected.Name, selected.ArtistName, selected.ArtworkURL, "mp3", outDir, debug)
       if err != nil {
           return fmt.Errorf("error downloading mp3: %w", err)
       }
       fmt.Printf("Done. Saved to %s\n", path)
   case "3":
       fmt.Print("Downloading MP4... ")
       path, err := DownloadTrack(selected.Name, selected.ArtistName, selected.ArtworkURL, "mp4", outDir, debug)
       if err != nil {
           return fmt.Errorf("error downloading mp4: %w", err)
       }
       fmt.Printf("Done. Saved to %s\n", path)
   default:
       for {
           fmt.Println("Invalid choice. Please enter a valid option (1-3, default 1):")
           fmt.Print("Enter choice (1-3, default 1): ")
           fmt.Scanln(&choice)
           if choice == "" || choice == "1" || choice == "2" || choice == "3" {
               break
           }
       }
       switch choice {
       case "", "1":
           if err := GetLinks(selected.URL); err != nil {
               return fmt.Errorf("error getting links: %w", err)
           }
       case "2":
           fmt.Print("Downloading MP3... ")
           path, err := DownloadTrack(selected.Name, selected.ArtistName, selected.ArtworkURL, "mp3", outDir, debug)
           if err != nil {
               return fmt.Errorf("error downloading mp3: %w", err)
           }
           fmt.Printf("Done. Saved to %s\n", path)
       case "3":
           fmt.Print("Downloading MP4... ")
           path, err := DownloadTrack(selected.Name, selected.ArtistName, selected.ArtworkURL, "mp4", outDir, debug)
           if err != nil {
               return fmt.Errorf("error downloading mp4: %w", err)
           }
           fmt.Printf("Done. Saved to %s\n", path)
       }
   }
   return nil
}

func RunOnboarding() error {
	config := &Config{}

	fmt.Println("\n========== Apple Music API Setup ==========")
	fmt.Println("To use the search feature, you need Apple Music API credentials.")
	fmt.Println("Follow these steps to get them:")
	fmt.Println("1. Sign in to your Apple Developer account at https://developer.apple.com")
	fmt.Println("2. Go to Certificates, Identifiers & Profiles")
	fmt.Println("3. Under Keys, create a new key with MusicKit enabled")
	fmt.Println("4. Note down the Key ID, Team ID, and download the private key (.p8) file")
	fmt.Println("\nYou'll need to enter these values below:")

	fmt.Print("\nTeam ID: ")
	fmt.Scanln(&config.TeamID)
	config.TeamID = strings.TrimSpace(config.TeamID)
	if config.TeamID == "" {
		return errors.New("team ID cannot be empty")
	}

	fmt.Print("Key ID: ")
	fmt.Scanln(&config.KeyID)
	config.KeyID = strings.TrimSpace(config.KeyID)
	if config.KeyID == "" {
		return errors.New("key ID cannot be empty")
	}

	fmt.Print("Music ID (usually same as Team ID): ")
	fmt.Scanln(&config.MusicID)
	config.MusicID = strings.TrimSpace(config.MusicID)
	if config.MusicID == "" {
		config.MusicID = config.TeamID
	}

	fmt.Println("\nPath to your .p8 private key file:")
	var keyPath string
	fmt.Scanln(&keyPath)
	keyPath = strings.TrimSpace(keyPath)
	if keyPath == "" {
		return errors.New("key path cannot be empty")
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	config.PrivateKey = string(keyData)
	config.PrivateKey = strings.TrimSpace(config.PrivateKey)
	if config.PrivateKey == "" {
		return errors.New("private key file is empty")
	}

	err = config.SaveConfig()
	if err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	fmt.Println("\n✅ Apple Music API credentials saved successfully!")
	return nil
}