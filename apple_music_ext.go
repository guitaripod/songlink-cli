package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/marcusziade/musickitkat/models"
)

// AppleMusicClient extends the Apple Music API functionality
type AppleMusicClient struct {
	BaseURL        string
	DeveloperToken string
	HTTPClient     *http.Client
}

// NewAppleMusicClient creates a new Apple Music API client
func NewAppleMusicClient(developerToken string) *AppleMusicClient {
	return &AppleMusicClient{
		BaseURL:        "https://api.music.apple.com/v1",
		DeveloperToken: developerToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to the Apple Music API
func (c *AppleMusicClient) doRequest(ctx context.Context, method, path string, params url.Values) (*http.Response, error) {
	url := c.BaseURL + path
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.DeveloperToken)
	req.Header.Set("Accept", "application/json")

	return c.HTTPClient.Do(req)
}

// GetAlbumTracks fetches all tracks for an album
func (c *AppleMusicClient) GetAlbumTracks(ctx context.Context, storefront, albumID string) ([]models.Song, error) {
	path := fmt.Sprintf("/catalog/%s/albums/%s/tracks", storefront, albumID)
	
	var allTracks []models.Song
	limit := 100
	offset := 0

	for {
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", limit))
		params.Set("offset", fmt.Sprintf("%d", offset))

		resp, err := c.doRequest(ctx, "GET", path, params)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		var response models.SongsResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allTracks = append(allTracks, response.Data...)

		// Check if there are more tracks
		if response.Next == "" || len(response.Data) < limit {
			break
		}

		offset += limit
	}

	return allTracks, nil
}

// GetPlaylistTracks fetches all tracks for a playlist
func (c *AppleMusicClient) GetPlaylistTracks(ctx context.Context, storefront, playlistID string) ([]models.Song, error) {
	// First get the playlist with tracks relationship
	path := fmt.Sprintf("/catalog/%s/playlists/%s", storefront, playlistID)
	params := url.Values{}
	params.Set("include", "tracks")

	resp, err := c.doRequest(ctx, "GET", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var playlistResp models.PlaylistsResponse
	if err := json.NewDecoder(resp.Body).Decode(&playlistResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(playlistResp.Data) == 0 {
		return nil, fmt.Errorf("playlist not found")
	}

	playlist := playlistResp.Data[0]

	// Extract track IDs from relationships
	var trackIDs []string
	if playlist.Relationships.Tracks.Data != nil {
		for _, track := range playlist.Relationships.Tracks.Data {
			trackIDs = append(trackIDs, track.ID)
		}
	}

	if len(trackIDs) == 0 {
		return nil, fmt.Errorf("no tracks found in playlist")
	}

	// Fetch track details in batches
	var allTracks []models.Song
	for i := 0; i < len(trackIDs); i += 100 {
		end := i + 100
		if end > len(trackIDs) {
			end = len(trackIDs)
		}

		batchIDs := trackIDs[i:end]
		path := fmt.Sprintf("/catalog/%s/songs", storefront)
		params := url.Values{}
		params.Set("ids", joinIDs(batchIDs))

		resp, err := c.doRequest(ctx, "GET", path, params)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch songs: %w", err)
		}
		defer resp.Body.Close()

		var songsResp models.SongsResponse
		if err := json.NewDecoder(resp.Body).Decode(&songsResp); err != nil {
			return nil, fmt.Errorf("failed to decode songs: %w", err)
		}

		allTracks = append(allTracks, songsResp.Data...)
	}

	return allTracks, nil
}

// joinIDs joins string IDs with commas
func joinIDs(ids []string) string {
	result := ""
	for i, id := range ids {
		if i > 0 {
			result += ","
		}
		result += id
	}
	return result
}