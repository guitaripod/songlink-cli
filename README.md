# Songlink CLI

[![Github All Releases](https://img.shields.io/github/downloads/guitaripod/songlink-cli/total.svg)](https://github.com/guitaripod/songlink-cli/releases)

A Go program that retrieves Songlink and Spotify links for a given URL using the Songlink API. It also allows searching for songs and albums directly using Apple Music API. The output is designed to be shared as is, allowing the receiver to both use Songlink and listen to the song preview using Spotify's embed feature.

## Features

-   Retrieves Songlink and Spotify links for a given song or album URL
-   Search for songs and albums directly using Apple Music API
-   Download full tracks as MP3 or MP4 files with album artwork
-   Download entire playlists or albums from Apple Music URLs
-   Supports command line arguments for customizing the output format
-   Automatically copies the output to the clipboard for easy sharing
-   Includes a loading indicator to provide visual feedback during the retrieval process
-   Thoroughly tested with unit tests to ensure reliability and correctness

## Installation

### macOS

#### Homebrew

```
brew tap guitaripod/songlink-cli
brew install songlink-cli
```

#### Build

1. Clone the repository: `git clone https://github.com/guitaripod/songlink-cli.git`
2. Navigate to the project directory: `cd songlink-cli`
3. Install dependencies: `go mod download`
4. Build the executable: `go build -o songlink .`
5. Run the program: `./songlink`

### Download and Run

Go to [Releases](https://github.com/guitaripod/songlink-cli/releases) and download the appropriate version for your operating system (Linux, macOS, Windows).

### Dependencies for Download Features

To use the download functionality (single tracks or playlists), you need:

- `yt-dlp` - For downloading audio from YouTube
- `ffmpeg` - For audio/video processing

Install on macOS:
```bash
brew install yt-dlp ffmpeg
```

Install on Linux:
```bash
# Ubuntu/Debian
sudo apt install yt-dlp ffmpeg

# Arch
sudo pacman -S yt-dlp ffmpeg
```

## Usage

### Process URL from clipboard

1. Copy the URL of the song or album you want to retrieve links for.
2. Run the program using one of the following commands:
    - `./songlink`: Retrieves only the Songlink URL
    - `./songlink -x`: Retrieves the Songlink URL without surrounding `<>`. For Twitter
    - `./songlink -d`: Retrieves the Songlink URL surrounded by `<>` and the Spotify URL. For Discord.
    - `./songlink -s`: Retrieves only the Spotify URL
3. The program will automatically retrieve the Songlink and/or Spotify link for the song or album and copy it to your clipboard.

### Search for songs or albums

1. Configure your Apple Music API credentials (first time only):
   ```
   ./songlink config
   ```
   
2. Search for a song or album:
   ```
   ./songlink search "song or album name"
   ```
   
3. Select from the search results by entering the number.

4. After selecting a result, you will be prompted to choose an action:
   1) Copy the song.link + Spotify URL to clipboard  
   2) Download the full track as MP3  
   3) Download a video (MP4) with the album artwork

5. If you choose to download, the file(s) will be saved in the `downloads/` directory by default.

#### Search Flags

- `-type=song`: Search for songs only (default)
- `-type=album`: Search for albums only
- `-type=both`: Search for both songs and albums

Combined with output format flags:
```
./songlink search -type=album -d "Dark Side of the Moon"
```

### Download full tracks

You can download the full track audio or a video with artwork.

```bash
./songlink download [flags] <query>
```

Flags:

- `-type=song` / `album` / `both` (default: song) — Type of Apple Music search.  
- `-format=mp3` / `mp4` (default: mp3) — Download as an audio file (MP3) or a video with artwork (MP4).  
- `-out=DIR` (default: downloads) — Directory to save the downloaded files.

Example:

```bash
./songlink download -type=song -format=mp4 "Purple Rain"
```

### Download entire playlists or albums

Download all tracks from an Apple Music playlist or album URL.

```bash
./songlink playlist [flags] <apple-music-url>
```

Flags:

- `--format=mp3` / `mp4` (default: mp3) — Download format for all tracks
- `--out=DIR` (default: downloads) — Output directory for downloaded files
- `--concurrent=N` (default: 3) — Number of parallel downloads
- `--metadata` — Save playlist/album metadata as JSON
- `--debug` — Show detailed download progress and debug info

Examples:

```bash
# Download an album
./songlink playlist "https://music.apple.com/us/album/abbey-road/401469823"

# Download a playlist with metadata
./songlink playlist --metadata "https://music.apple.com/us/playlist/top-100-global/pl.d25f5d1181894928af76c85c967f8f31"

# Download with custom settings
./songlink playlist --format=mp4 --out=my-music --concurrent=5 "https://music.apple.com/album/..."
```

The playlist download feature:
- Downloads all tracks in parallel for faster completion
- Automatically retries failed downloads
- Shows progress for each track
- Saves metadata including track order and download status
- Creates the output directory if it doesn't exist

## Apple Music API Setup

To use the search functionality, you need Apple Music API credentials. The CLI includes a guided setup process:

1. Run `./songlink config`
2. Follow the prompts to enter your Apple Developer credentials:
   - Team ID
   - Key ID
   - Private Key (from your .p8 file)
   - Music ID (usually the same as Team ID)

Your credentials will be securely stored in `~/.songlink-cli/config.json`

## Examples

Here are a few examples of how to use the Songlink CLI:

-   Retrieve only the Songlink URL:

    ```
    ./songlink
    ```

-   Retrieve the Songlink URL without surrounding `<>` + Spotify embed:

    ```
    ./songlink -x
    ```

-   Retrieve the Songlink URL surrounded by `<>` + Spotify embed:

    ```
    ./songlink -d
    ```

-   Retrieve only the Spotify URL:
    ```
    ./songlink -s
    ```

-   Search for a song and get links:
    ```
    ./songlink search "Bohemian Rhapsody"
    ```

-   Search for an album with specific output format:
    ```
    ./songlink search -type=album -d "Abbey Road"
    ```

-   Download an entire album from Apple Music:
    ```
    ./songlink playlist "https://music.apple.com/gb/album/final-fantasy-x-lofi-sound-of-spira/1693470715"
    ```

-   Download a playlist with metadata:
    ```
    ./songlink playlist --metadata --out=playlists "https://music.apple.com/us/playlist/..."
    ```

## Contributions

I welcome contributions to the Songlink CLI project! If you have any ideas, suggestions, or bug reports, please don't hesitate to open an issue or submit a pull request. To contribute:

1. Fork the repository
2. Create a new branch for your feature or bug fix
3. Make your changes and commit them with descriptive commit messages
4. Push your changes to your forked repository
5. Submit a pull request to the main repository

I appreciate your help in making this project better!

## License

This project is licensed under the [MIT License](LICENSE).

---

I hope you find this tool useful! If you have any questions or need further assistance, please let me know.
