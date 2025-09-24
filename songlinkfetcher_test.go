package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeRequest(t *testing.T) {
	searchURL := "https://music.apple.com/fi/album/caravan/1572919347?i=1572919354"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"pageUrl": "https://song.link/fi/i/1572919354", "linksByPlatform": {"spotify": {"url": "https://open.spotify.com/track/2Xtsv7BUMrNodQWH2JPOc0"}}}`)
	}))
	defer server.Close()

	response, err := makeRequest(searchURL)

	if err != nil {
		t.Errorf("makeRequest(%q) returned an unexpected error: %v", searchURL, err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("makeRequest(%q) returned a non-OK HTTP response status: %s", searchURL, response.Status)
	}

	linksResponse := SonglinkResponse{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&linksResponse)
	if err != nil {
		t.Errorf("makeRequest(%q) returned an invalid JSON response body: %v", searchURL, err)
	}

	if linksResponse.PageURL != "https://song.link/fi/i/1572919354" {
		t.Errorf("makeRequest(%q) returned an unexpected page URL: %s", searchURL, linksResponse.PageURL)
	}

   expectedSpotifyURL := "https://open.spotify.com/track/2Xtsv7BUMrNodQWH2JPOc0"
   if linksResponse.LinksByPlatform.Spotify.URL != expectedSpotifyURL {
       t.Errorf("makeRequest(%q) returned an unexpected Spotify URL: %s (want %s)", searchURL, linksResponse.LinksByPlatform.Spotify.URL, expectedSpotifyURL)
   }
}

func TestBuildURL(t *testing.T) {
	searchURL := "https://music.apple.com/fi/album/caravan/1572919347?i=1572919354"
	expectedURL := "https://api.song.link/v1-alpha.1/links?url=https%3A%2F%2Fmusic.apple.com%2Ffi%2Falbum%2Fcaravan%2F1572919347%3Fi%3D1572919354"
	actualURL := buildURL(searchURL)
	if actualURL != expectedURL {
		t.Errorf("buildURL(%q) = %q; want %q", searchURL, actualURL, expectedURL)
	}
}
