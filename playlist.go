package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// PlaylistURLParser handles parsing of Apple Music URLs
type PlaylistURLParser struct {
	playlistPattern *regexp.Regexp
	albumPattern    *regexp.Regexp
}

// NewPlaylistURLParser creates a new URL parser
func NewPlaylistURLParser() *PlaylistURLParser {
	return &PlaylistURLParser{
		// Match playlist URLs: https://music.apple.com/{country}/playlist/{name}/pl.{id}
		playlistPattern: regexp.MustCompile(`music\.apple\.com/([a-z]{2})/playlist/[^/]+/pl\.([a-zA-Z0-9]+)`),
		// Match album URLs: https://music.apple.com/{country}/album/{name}/{id}
		albumPattern: regexp.MustCompile(`music\.apple\.com/([a-z]{2})/album/[^/]+/(\d+)`),
	}
}

// ParseResourceType represents the type of parsed resource
type ParseResourceType string

const (
	ParsedPlaylist ParseResourceType = "playlist"
	ParsedAlbum    ParseResourceType = "album"
)

// ParsedResource contains the parsed URL information
type ParsedResource struct {
	Type       ParseResourceType
	ID         string
	Storefront string // Country code (e.g., "us", "gb")
}

// Parse extracts resource type and ID from an Apple Music URL
func (p *PlaylistURLParser) Parse(inputURL string) (*ParsedResource, error) {
	// Validate URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure it's an Apple Music URL
	if !strings.Contains(parsedURL.Host, "music.apple.com") {
		return nil, errors.New("not an Apple Music URL")
	}

	// Try to match playlist pattern
	if matches := p.playlistPattern.FindStringSubmatch(inputURL); len(matches) > 2 {
		return &ParsedResource{
			Type:       ParsedPlaylist,
			ID:         matches[2],
			Storefront: matches[1],
		}, nil
	}

	// Try to match album pattern
	if matches := p.albumPattern.FindStringSubmatch(inputURL); len(matches) > 2 {
		return &ParsedResource{
			Type:       ParsedAlbum,
			ID:         matches[2],
			Storefront: matches[1],
		}, nil
	}

	return nil, errors.New("unrecognized Apple Music URL format")
}

// ExtendedMusicSearcher extends MusicSearcher with playlist/album track fetching
type ExtendedMusicSearcher struct {
	*MusicSearcher
}

// NewExtendedMusicSearcher creates a new extended music searcher
func NewExtendedMusicSearcher(config *Config) (*ExtendedMusicSearcher, error) {
	searcher, err := NewMusicSearcher(config)
	if err != nil {
		return nil, err
	}
	return &ExtendedMusicSearcher{MusicSearcher: searcher}, nil
}

// GetAlbumWithTracks fetches an album with all its tracks
func (ems *ExtendedMusicSearcher) GetAlbumWithTracks(ctx context.Context, albumID string, storefront string) (*AlbumWithTracks, error) {
	// Set the storefront for the catalog service
	ems.client.Catalog.SetStorefront(storefront)

	// Fetch the album details
	album, err := ems.client.Catalog.GetAlbum(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	// Create an extended Apple Music client to fetch tracks
	apiClient := NewAppleMusicClient(ems.client.DeveloperToken)
	
	// Get all tracks for this album
	songs, err := apiClient.GetAlbumTracks(ctx, storefront, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tracks: %w", err)
	}

	// Convert to SearchResult format
	var tracks []SearchResult
	for _, song := range songs {
		artURL := song.Attributes.Artwork.URL
		artURL = strings.ReplaceAll(artURL, "{w}", "500")
		artURL = strings.ReplaceAll(artURL, "{h}", "500")
		
		tracks = append(tracks, SearchResult{
			ID:         song.ID,
			Name:       song.Attributes.Name,
			ArtistName: song.Attributes.ArtistName,
			Type:       Song,
			URL:        song.Attributes.URL,
			ArtworkURL: artURL,
		})
	}

	// Build artwork URL
	artURL := album.Attributes.Artwork.URL
	artURL = strings.ReplaceAll(artURL, "{w}", "500")
	artURL = strings.ReplaceAll(artURL, "{h}", "500")

	return &AlbumWithTracks{
		ID:         album.ID,
		Name:       album.Attributes.Name,
		ArtistName: album.Attributes.ArtistName,
		ArtworkURL: artURL,
		Tracks:     tracks,
		TrackCount: album.Attributes.TrackCount,
	}, nil
}

// GetPlaylistWithTracks fetches a playlist with all its tracks
func (ems *ExtendedMusicSearcher) GetPlaylistWithTracks(ctx context.Context, playlistID string, storefront string) (*PlaylistWithTracks, error) {
	// Set the storefront for the catalog service
	ems.client.Catalog.SetStorefront(storefront)

	// Fetch playlist details
	playlist, err := ems.client.Catalog.GetPlaylist(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}

	// Create an extended Apple Music client to fetch tracks
	apiClient := NewAppleMusicClient(ems.client.DeveloperToken)
	
	// Get all tracks for this playlist
	songs, err := apiClient.GetPlaylistTracks(ctx, storefront, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}

	// Convert to SearchResult format
	var tracks []SearchResult
	for _, song := range songs {
		artURL := song.Attributes.Artwork.URL
		artURL = strings.ReplaceAll(artURL, "{w}", "500")
		artURL = strings.ReplaceAll(artURL, "{h}", "500")
		
		tracks = append(tracks, SearchResult{
			ID:         song.ID,
			Name:       song.Attributes.Name,
			ArtistName: song.Attributes.ArtistName,
			Type:       Song,
			URL:        song.Attributes.URL,
			ArtworkURL: artURL,
		})
	}

	// Build artwork URL
	artURL := ""
	if playlist.Attributes.Artwork.URL != "" {
		artURL = playlist.Attributes.Artwork.URL
		artURL = strings.ReplaceAll(artURL, "{w}", "500")
		artURL = strings.ReplaceAll(artURL, "{h}", "500")
	}

	return &PlaylistWithTracks{
		ID:          playlist.ID,
		Name:        playlist.Attributes.Name,
		CuratorName: playlist.Attributes.CuratorName,
		ArtworkURL:  artURL,
		Tracks:      tracks,
		TrackCount:  playlist.Attributes.TrackCount,
	}, nil
}

// AlbumWithTracks represents an album with its tracks
type AlbumWithTracks struct {
	ID         string
	Name       string
	ArtistName string
	ArtworkURL string
	Tracks     []SearchResult
	TrackCount int
}

// PlaylistWithTracks represents a playlist with its tracks
type PlaylistWithTracks struct {
	ID          string
	Name        string
	CuratorName string
	ArtworkURL  string
	Tracks      []SearchResult
	TrackCount  int
}