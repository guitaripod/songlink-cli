# Songlink CLI

[![Github All Releases](https://img.shields.io/github/downloads/guitaripod/songlink-cli/total.svg)](https://github.com/guitaripod/songlink-cli/releases)

A Go program that retrieves Songlink and Spotify links for a given URL using the Songlink API. It also allows searching for songs and albums directly using Apple Music API. The output is designed to be shared as is, allowing the receiver to both use Songlink and listen to the song preview using Spotify's embed feature.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Process URL from clipboard](#process-url-from-clipboard)
  - [Search for songs or albums](#search-for-songs-or-albums)
  - [Download single tracks](#download-single-tracks)
  - [Download playlists/albums](#download-entire-playlists-or-albums)
- [Examples](#examples)
- [Contributions](#contributions)
- [License](#license)

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

<details>
<summary><strong>📦 Installation Methods</strong></summary>

### macOS

#### Homebrew

```
brew tap guitaripod/songlink-cli
brew install songlink-cli
```

#### Build from Source

1. Clone the repository: `git clone https://github.com/guitaripod/songlink-cli.git`
2. Navigate to the project directory: `cd songlink-cli`
3. Install dependencies: `go mod download`
4. Build the executable: `go build -o songlink .`
5. Run the program: `./songlink`

### Download Pre-built Binaries

Go to [Releases](https://github.com/guitaripod/songlink-cli/releases) and download the appropriate version for your operating system (Linux, macOS, Windows).

</details>

<details>
<summary><strong>🛠️ Dependencies for Download Features</strong></summary>

To use the download functionality (single tracks or playlists), you need:

- `yt-dlp` - For downloading audio from YouTube
- `ffmpeg` - For audio/video processing

**Install on macOS:**
```bash
brew install yt-dlp ffmpeg
```

**Install on Linux:**
```bash
# Ubuntu/Debian
sudo apt install yt-dlp ffmpeg

# Arch
sudo pacman -S yt-dlp ffmpeg
```

</details>

## Usage

<details>
<summary><strong>📋 Process URL from Clipboard</strong></summary>

1. Copy the URL of the song or album you want to retrieve links for.
2. Run the program using one of the following commands:
    - `./songlink`: Retrieves only the Songlink URL
    - `./songlink -x`: Retrieves the Songlink URL without surrounding `<>`. For Twitter
    - `./songlink -d`: Retrieves the Songlink URL surrounded by `<>` and the Spotify URL. For Discord.
    - `./songlink -s`: Retrieves only the Spotify URL
3. The program will automatically retrieve the Songlink and/or Spotify link for the song or album and copy it to your clipboard.

</details>

<details>
<summary><strong>🔍 Search for Songs or Albums</strong></summary>

### Basic Search

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

### Search Flags

- `-type=song`: Search for songs only (default)
- `-type=album`: Search for albums only
- `-type=both`: Search for both songs and albums

Combined with output format flags:
```
./songlink search -type=album -d "Dark Side of the Moon"
```

</details>

<details>
<summary><strong>💿 Download Single Tracks</strong></summary>

Download individual tracks as audio files or videos with artwork.

```bash
./songlink download [flags] <query>
```

### Flags

- `-type=song` / `album` / `both` (default: song) — Type of Apple Music search.  
- `-format=mp3` / `mp4` (default: mp3) — Download as an audio file (MP3) or a video with artwork (MP4).  
- `-out=DIR` (default: downloads) — Directory to save the downloaded files.

### Example

```bash
./songlink download -type=song -format=mp4 "Purple Rain"
```

</details>

<details>
<summary><strong>📀 Download Entire Playlists or Albums</strong></summary>

Download all tracks from an Apple Music playlist or album URL.

```bash
./songlink playlist [flags] <apple-music-url>
```

### ⚠️ Supported Content

**✅ Supported:**
- Public catalog albums (e.g., `https://music.apple.com/us/album/album-name/123456789`)
- Public catalog playlists (e.g., `https://music.apple.com/us/playlist/playlist-name/pl.abcdef123456`)

**❌ Not Supported:**
- Personal library playlists (`/library/playlist/`)
- Private or user-created playlists
- Region-locked content (may return 404 errors)
- Apple Music Radio stations

### Flags

- `--format=mp3` / `mp4` (default: mp3) — Download format for all tracks
- `--out=DIR` (default: downloads) — Output directory for downloaded files
- `--concurrent=N` (default: 3) — Number of parallel downloads
- `--metadata` — Save playlist/album metadata as JSON
- `--debug` — Show detailed download progress and debug info

### Examples

```bash
# Download an album
./songlink playlist "https://music.apple.com/us/album/abbey-road/401469823"

# Download a playlist with metadata
./songlink playlist --metadata "https://music.apple.com/us/playlist/top-100-global/pl.d25f5d1181894928af76c85c967f8f31"

# Download with custom settings
./songlink playlist --format=mp4 --out=my-music --concurrent=5 "https://music.apple.com/album/..."
```

### Features

- **Parallel Downloads**: Downloads multiple tracks simultaneously for faster completion
- **Automatic Retry**: Failed downloads are retried with exponential backoff
- **Progress Tracking**: Real-time progress for each track
- **Metadata Support**: Saves playlist/album info and track details as JSON
- **Smart Directory Creation**: Automatically creates output directories

### Troubleshooting

If you get a "404 Resource Not Found" error:
- The playlist/album may be region-locked
- The content may have been removed
- Try using a different storefront in the URL (e.g., `/us/`, `/gb/`, `/jp/`)

</details>

<details>
<summary><strong>🔐 Apple Music API Setup</strong></summary>

To use the search and download functionality, you need Apple Music API credentials. The CLI includes a guided setup process:

1. Run `./songlink config`
2. Follow the prompts to enter your Apple Developer credentials:
   - Team ID
   - Key ID
   - Private Key (from your .p8 file)
   - Music ID (usually the same as Team ID)

Your credentials will be securely stored in `~/.songlink-cli/config.json`

### Getting Apple Music API Credentials

1. Sign in to [Apple Developer](https://developer.apple.com)
2. Go to Certificates, Identifiers & Profiles
3. Under Keys, create a new key with MusicKit enabled
4. Download the .p8 private key file
5. Note your Team ID and Key ID

</details>

## Examples

<details>
<summary><strong>📚 Quick Examples</strong></summary>

### Link Retrieval

```bash
# Get Songlink URL (default)
./songlink

# Get Songlink URL without <> for Twitter
./songlink -x

# Get Songlink URL with <> + Spotify URL for Discord
./songlink -d

# Get only Spotify URL
./songlink -s
```

### Search Examples

```bash
# Search for a song
./songlink search "Bohemian Rhapsody"

# Search for an album with Discord format
./songlink search -type=album -d "Abbey Road"

# Search for both songs and albums
./songlink search -type=both "Beatles"
```

### Download Examples

```bash
# Download a single track as MP3
./songlink download "Purple Rain"

# Download a single track as MP4 with artwork
./songlink download -format=mp4 "Imagine"

# Download an entire album
./songlink playlist "https://music.apple.com/us/album/abbey-road/401469823"

# Download album with metadata and custom settings
./songlink playlist --metadata --out=my-music --concurrent=5 "https://music.apple.com/us/album/..."
```

</details>

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
