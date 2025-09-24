package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type PlaylistURLParser struct {
	playlistPattern *regexp.Regexp
	albumPattern    *regexp.Regexp
}

func NewPlaylistURLParser() *PlaylistURLParser {
	return &PlaylistURLParser{
		playlistPattern: regexp.MustCompile(`music\.apple\.com/([a-z]{2})/playlist/[^/]+/pl\.([a-zA-Z0-9]+)`),
		albumPattern: regexp.MustCompile(`music\.apple\.com/([a-z]{2})/album/[^/]+/(\d+)`),
	}
}

type ParseResourceType string

const (
	ParsedPlaylist ParseResourceType = "playlist"
	ParsedAlbum    ParseResourceType = "album"
)

type ParsedResource struct {
	Type       ParseResourceType
	ID         string
	Storefront string
}

func (p *PlaylistURLParser) Parse(inputURL string) (*ParsedResource, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if !strings.Contains(parsedURL.Host, "music.apple.com") {
		return nil, errors.New("not an Apple Music URL")
	}

	if matches := p.playlistPattern.FindStringSubmatch(inputURL); len(matches) > 2 {
		return &ParsedResource{
			Type:       ParsedPlaylist,
			ID:         matches[2],
			Storefront: matches[1],
		}, nil
	}

	if matches := p.albumPattern.FindStringSubmatch(inputURL); len(matches) > 2 {
		return &ParsedResource{
			Type:       ParsedAlbum,
			ID:         matches[2],
			Storefront: matches[1],
		}, nil
	}

	return nil, errors.New("unrecognized Apple Music URL format")
}

type ExtendedMusicSearcher struct {
	*MusicSearcher
}

func NewExtendedMusicSearcher(config *Config) (*ExtendedMusicSearcher, error) {
	searcher, err := NewMusicSearcher(config)
	if err != nil {
		return nil, err
	}
	return &ExtendedMusicSearcher{MusicSearcher: searcher}, nil
}

func (ems *ExtendedMusicSearcher) GetAlbumWithTracks(ctx context.Context, albumID string, storefront string) (*AlbumWithTracks, error) {
	ems.client.Catalog.SetStorefront(storefront)

	album, err := ems.client.Catalog.GetAlbum(ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	apiClient := NewAppleMusicClient(ems.client.DeveloperToken)
	
	songs, err := apiClient.GetAlbumTracks(ctx, storefront, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album tracks: %w", err)
	}

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

func (ems *ExtendedMusicSearcher) GetPlaylistWithTracks(ctx context.Context, playlistID string, storefront string) (*PlaylistWithTracks, error) {
	apiClient := NewAppleMusicClient(ems.client.DeveloperToken)

	playlist, err := apiClient.GetPlaylistDetails(ctx, storefront, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	
	songs, err := apiClient.GetPlaylistTracks(ctx, storefront, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}

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

type AlbumWithTracks struct {
	ID         string
	Name       string
	ArtistName string
	ArtworkURL string
	Tracks     []SearchResult
	TrackCount int
}

type PlaylistWithTracks struct {
	ID          string
	Name        string
	CuratorName string
	ArtworkURL  string
	Tracks      []SearchResult
	TrackCount  int
}