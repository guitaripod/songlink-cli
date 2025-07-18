package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/atotto/clipboard"
)

// SonglinkResponse represents the response from the song.link API.
type SonglinkResponse struct {
	PageURL         string          `json:"pageUrl"`
	LinksByPlatform LinksByPlatform `json:"linksByPlatform"`
}

// LinksByPlatform contains links to the content on various music platforms.
type LinksByPlatform struct {
	Spotify PlatformMusic `json:"spotify"`
}

// PlatformMusic represents a single platform's music link.
type PlatformMusic struct {
	URL string `json:"url"`
}

// GetLinks fetches shareable links for a given music URL and copies them to the clipboard.
// Formats the output based on command-line flags (-x, -d, -s).
func GetLinks(searchURL string) error {
	response, err := makeRequest(searchURL)
	if err != nil {
		return err
	}

	var linksResponse SonglinkResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&linksResponse)
	if err != nil {
		return fmt.Errorf("error decoding JSON response: %w", err)
	}

	nonLocalURL := strings.ReplaceAll(linksResponse.PageURL, "/fi", "")
	spotifyURL := linksResponse.LinksByPlatform.Spotify.URL

	var outputString string
	if *xFlag {
		outputString = fmt.Sprintf("%s\n%s", nonLocalURL, spotifyURL)
	} else if *dFlag {
		outputString = fmt.Sprintf("<%s>\n%s", nonLocalURL, spotifyURL)
	} else if *sFlag {
		outputString = spotifyURL
	} else {
		outputString = nonLocalURL
	}

	err = clipboard.WriteAll(outputString)
	if err != nil {
		return fmt.Errorf("error copying output string to clipboard: %w", err)
	}

	fmt.Print(
		"\nSuccess ✅\n",
		outputString,
		"\nCopied to the clipboard\n\n",
	)

	return nil
}

// makeRequest performs an HTTP GET request to the song.link API.
func makeRequest(searchURL string) (*http.Response, error) {
	url := buildURL(searchURL)
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP response status: %s", response.Status)
	}

	return response, nil
}

// buildURL constructs the song.link API URL with the search URL as a parameter.
func buildURL(searchURL string) string {
	url := url.URL{
		Scheme: "https",
		Host:   "api.song.link",
		Path:   "/v1-alpha.1/links",
	}
	values := url.Query()
	values.Add("url", searchURL)
	url.RawQuery = values.Encode()
	return url.String()
}
