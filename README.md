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
-   Comprehensive built-in help system (`songlink-cli --help`)
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
4. Build the executable: `go build -o songlink-cli .`
5. Run the program: `./songlink-cli`

### Download Pre-built Binaries

Go to [Releases](https://github.com/guitaripod/songlink-cli/releases) and download the appropriate version for your operating system (Linux, macOS, Windows).

</details>

<details>
<summary><strong>🛠️ Dependencies for Download Features</strong></summary>

### ⚠️ IMPORTANT: yt-dlp Version Requirements

To use the download functionality (single tracks or playlists), you need:

- **`yt-dlp`** - For downloading audio from YouTube **(MUST be latest version)**
- **`ffmpeg`** - For audio/video processing

> **⚠️ Critical:** YouTube frequently changes their API, breaking older versions of yt-dlp. You **MUST** keep yt-dlp updated to the latest version or downloads will fail.

### Installing yt-dlp (Latest Version)

**macOS (Homebrew):**
```bash
# Install and keep updated
brew install yt-dlp ffmpeg
brew upgrade yt-dlp  # Run regularly to stay updated
```

**Linux (Recommended methods for latest version):**

```bash
# Option 1: Direct download (always latest)
curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o ~/.local/bin/yt-dlp
chmod +x ~/.local/bin/yt-dlp
# Add ~/.local/bin to your PATH if not already there

# Option 2: pip/pipx (easy updates)
pipx install yt-dlp  # or: pip install --user yt-dlp
pipx upgrade yt-dlp  # Run regularly to stay updated

# Option 3: Package manager (may be outdated)
# Ubuntu/Debian
sudo apt install yt-dlp ffmpeg
# Note: apt version is often outdated. Consider Option 1 or 2 instead.

# Arch Linux (AUR version stays current)
yay -S yt-dlp-git ffmpeg  # yt-dlp-git auto-updates
# Or for manual updates:
yay -S yt-dlp ffmpeg
yay -Syu yt-dlp  # Run regularly
```

**Verify yt-dlp is up to date:**
```bash
yt-dlp --version
# Should show a recent date (within last month)
# If older, update immediately
```

### Troubleshooting Download Errors

If downloads fail with errors like:
- "Requested format is not available"
- "Signature extraction failed"
- "Some formats may be missing"

**This means yt-dlp is outdated.** Update it immediately:
```bash
# Quick update based on your system:
yt-dlp -U  # If installed via direct download
pipx upgrade yt-dlp  # If installed via pipx
brew upgrade yt-dlp  # macOS
yay -Syu yt-dlp  # Arch Linux
```

</details>

## Usage

```
songlink-cli [flags]                    Process URL from clipboard
songlink-cli <command> [flags] <args>   Run a specific command  
songlink-cli help <command>             Show help for a command
```

Run `songlink-cli --help` for comprehensive documentation of all features.

<details>
<summary><strong>📋 Process URL from Clipboard</strong></summary>

1. Copy the URL of the song or album you want to retrieve links for.
2. Run the program using one of the following commands:
    - `songlink-cli`: Retrieves only the Songlink URL
    - `songlink-cli -x`: Retrieves the Songlink URL without surrounding `<>` (for Twitter)
    - `songlink-cli -d`: Retrieves the Songlink URL surrounded by `<>` and the Spotify URL (for Discord)
    - `songlink-cli -s`: Retrieves only the Spotify URL
3. The program will automatically retrieve the Songlink and/or Spotify link for the song or album and copy it to your clipboard.

</details>

<details>
<summary><strong>🔍 Search for Songs or Albums</strong></summary>

### Basic Search

1. Configure your Apple Music API credentials (first time only):
   ```bash
   songlink-cli config
   ```
   
2. Search for a song or album:
   ```bash
   songlink-cli search "song or album name"
   ```
   
3. Select from the search results by entering the number.

4. After selecting a result, you will be prompted to choose an action:
   - **Option 1**: Copy the song.link + Spotify URL to clipboard  
   - **Option 2**: Download the full track as MP3  
   - **Option 3**: Download a video (MP4) with the album artwork

5. If you choose to download, the file(s) will be saved in the `downloads/` directory by default.

### Advanced Search Examples

```bash
# Search for a specific song
songlink-cli search "Bohemian Rhapsody"

# Search for albums only
songlink-cli search -type=album "Abbey Road"

# Search and format for Discord sharing
songlink-cli search -d "Hotel California"

# Search with custom download directory
songlink-cli search -out=~/Music "Imagine"
```

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

| Flag | Options | Default | Description |
|------|---------|---------|-------------|
| `-type` | `song`, `album` | `song` | Type of Apple Music search |
| `-format` | `mp3`, `mp4` | `mp3` | Download format (MP3 audio or MP4 video with artwork) |
| `-out` | Directory path | `downloads` | Output directory for downloaded files |
| `-debug` | - | `false` | Show yt-dlp and ffmpeg output |

### Examples

```bash
# Download a song as MP3 (default)
songlink-cli download "Stairway to Heaven"

# Download as MP4 with album artwork
songlink-cli download -format=mp4 "Purple Rain"

# Download to custom directory
songlink-cli download -out=~/Music "Yesterday"

# Download with debug output
songlink-cli download -debug "Wonderwall"
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

| Flag | Options | Default | Description |
|------|---------|---------|-------------|
| `--format` | `mp3`, `mp4` | `mp3` | Download format for all tracks |
| `--out` | Directory path | `downloads` | Output directory for downloaded files |
| `--concurrent` | `1-10` | `3` | Number of parallel downloads |
| `--metadata` | - | `false` | Save playlist/album metadata as JSON |
| `--debug` | - | `false` | Show detailed download progress and debug info |

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

#### 404 Resource Not Found Errors
- The playlist/album may be region-locked
- The content may have been removed
- Try using a different storefront in the URL (e.g., `/us/`, `/gb/`, `/jp/`)

#### Download Failures

**Most common cause: Outdated yt-dlp**
```bash
# Check your yt-dlp version
yt-dlp --version

# Update immediately if older than a month
yt-dlp -U  # or use your package manager
```

**Other causes:**
- Some tracks may not be available on YouTube
- Try reducing `--concurrent` value if experiencing rate limits
- Use `--debug` flag to see detailed error messages
- Ensure both `yt-dlp` AND `ffmpeg` are installed

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

### Link Retrieval from Clipboard

```bash
# Get Songlink URL (default)
songlink-cli

# Get Songlink URL without <> for Twitter
songlink-cli -x

# Get Songlink URL with <> + Spotify URL for Discord
songlink-cli -d

# Get only Spotify URL
songlink-cli -s
```

### Search Examples

```bash
# Search for a song
songlink-cli search "Bohemian Rhapsody"

# Search for an album with Discord format
songlink-cli search -type=album -d "Abbey Road"

# Search for both songs and albums
songlink-cli search -type=both "Beatles"

# Search with custom download directory
songlink-cli search -out=~/Downloads "Imagine"
```

### Download Examples

```bash
# Download a single track as MP3
songlink-cli download "Purple Rain"

# Download a single track as MP4 with artwork
songlink-cli download -format=mp4 "Imagine"

# Download to specific directory with debug output
songlink-cli download -out=~/Music -debug "Let It Be"
```

### Playlist/Album Download Examples

```bash
# Download an entire album
songlink-cli playlist "https://music.apple.com/us/album/abbey-road/401469823"

# Download playlist with metadata
songlink-cli playlist --metadata "https://music.apple.com/us/playlist/top-100-global/pl.d25f5d1181894928af76c85c967f8f31"

# Fast download with 5 workers and MP4 format
songlink-cli playlist --concurrent=5 --format=mp4 "https://music.apple.com/album/..."

# Download with all options
songlink-cli playlist --format=mp4 --out=my-music --concurrent=5 --metadata --debug "https://music.apple.com/playlist/..."
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
