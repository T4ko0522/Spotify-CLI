<h1 align="center">Spotify-CLI</h1>

<h3 align="center">Control Spotify from your terminal. TUI + CLI.</h3>

<p align="center">
  <img src="assets/readme.png" alt="spt" width="600" />
</p>

<p align="center">
  <a href="https://github.com/T4ko0522/Spotify-CLI/releases"><img src="https://img.shields.io/github/v/release/T4ko0522/Spotify-CLI?style=flat-square&label=version" alt="Release" /></a>
  <a href="https://github.com/T4ko0522/Spotify-CLI/blob/main/LICENSE"><img src="https://img.shields.io/github/license/T4ko0522/Spotify-CLI?style=flat-square" alt="License" /></a>
  <img src="https://img.shields.io/badge/go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.25+" />
  <img src="https://img.shields.io/badge/platform-windows%20%7C%20linux-lightgrey?style=flat-square" alt="Platform" />
</p>

<p align="center">
  Now Playing TUI · Album Art · Synced Lyrics · Volume Control · Device Management
</p>

<p align="center">
  English | <a href="README.ja.md">日本語</a>
</p>

---

## ✨ Features

- **Now Playing TUI** — real-time track info with album art (WezTerm)
- **Synced Lyrics** — timestamped lyrics via [LRCLIB](https://lrclib.net)
- **Playback Control** — play, stop, next, back
- **Volume Control** — TUI slider or direct value
- **Device Management** — list and switch active devices
- **Image Presets** — small / medium / large album art

## 📋 Requirements

- Client ID from [Spotify Developer](https://developer.spotify.com/)
- Redirect URI: `http://127.0.0.1:8888/callback`

## 📦 Installation

### Windows

**Scoop** (recommended — auto-update via `scoop update`):

```powershell
scoop bucket add t4ko0522 https://github.com/T4ko0522/scoop-bucket
scoop install t4ko0522/spt
```

**MSI installer**: Download `spt.msi` from [Releases](https://github.com/T4ko0522/Spotify-CLI/releases) and run it.

### Linux

**Homebrew** (recommended — auto-update via `brew upgrade`):

```bash
brew tap t4ko0522/tap https://github.com/T4ko0522/homebrew-tap
brew install spt
```

**From source** (requires Go 1.25+):

```bash
go install github.com/T4ko0522/spotify-cli@latest
```

**From binary**: Download from [Releases](https://github.com/T4ko0522/Spotify-CLI/releases) and place in your `$PATH`.

## 🚀 Quick Start

```bash
spt init     # Configure Client ID & authenticate
spt          # Launch Now Playing TUI
spt -l       # Show synced lyrics
```

## 📖 Commands

| Command | Alias | Description |
|---|---|---|
| `spt` | | Launch TUI (Now Playing) |
| `spt --lyrics` | `spt -l` | Show synced lyrics |
| `spt init` | | Configure Client ID & Spotify auth |
| `spt play` | `spt p` | Resume playback |
| `spt stop` | `spt s` | Pause |
| `spt next` | `spt n` | Next track |
| `spt back` | `spt b` | Previous track |
| `spt now` | | Show currently playing track |
| `spt volume` | `spt v` | Launch volume control TUI |
| `spt volume [0-100]` | `spt v [0-100]` | Set volume |
| `spt devices` | `spt d` | List available devices |
| `spt settings` | | Change image size preset |

## ⚙️ Settings

Use `spt settings` to change album art size. Arrow keys to select, Enter to confirm.

| Preset | Size |
|---|---|
| small | 16×8 |
| medium | 20×10 (default) |
| large | 28×14 |

## 📄 License

[Apache-2.0 LICENSE](https://github.com/T4ko0522/Spotify-CLI/blob/main/LICENSE)
