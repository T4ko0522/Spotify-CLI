<h1 align="center">Spotify-CLI</h1>

<h3 align="center">ターミナルからSpotifyを操作できます。TUI + CLI</h3>

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
  Now Playing TUI · アルバムアート · 同期歌詞 · 音量調整 · デバイス管理
</p>

<p align="center">
  <a href="README.md">English</a> | 日本語
</p>

---

## ✨ 機能

- **Now Playing TUI** — リアルタイムのトラック情報とアルバムアート表示（WezTerm）
- **同期歌詞** — [LRCLIB](https://lrclib.net) 経由のタイムスタンプ付き歌詞
- **再生制御** — play / stop / next / back
- **音量調整** — TUI スライダーまたは直接値指定
- **デバイス管理** — デバイス一覧とアクティブデバイス切替
- **画像プリセット** — small / medium / large のアルバムアートサイズ

## 📋 必要要件

- [Spotify Developer](https://developer.spotify.com/) で作成したアプリの Client ID
- リダイレクト URI: `http://127.0.0.1:8888/callback`

## 📦 インストール

### Windows

**Scoop**（推奨 — `scoop update` で自動更新）:

```powershell
scoop bucket add t4ko0522 https://github.com/T4ko0522/tap
scoop install t4ko0522/spt
```

**MSI インストーラー**: [Release](https://github.com/T4ko0522/Spotify-CLI/releases) から `spt.msi` をダウンロードして実行してください。

### Linux

**Homebrew**（推奨 — `brew upgrade` で自動更新）:

```bash
brew tap t4ko0522/tap https://github.com/T4ko0522/tap
brew install spt
```

**ソースからビルド**（Go 1.25+ が必要）:

```bash
go install github.com/T4ko0522/spotify-cli@latest
```

**バイナリ**: [Release](https://github.com/T4ko0522/Spotify-CLI/releases) からダウンロードし、`$PATH` に配置してください。

## 🚀 クイックスタート

```bash
spt init     # Client ID 設定 & Spotify 認証
spt          # Now Playing TUI を起動
spt -l       # 同期歌詞を表示
```

## 📖 コマンド一覧

| コマンド | エイリアス | 説明 |
|---|---|---|
| `spt` | | TUI を起動（Now Playing） |
| `spt --lyrics` | `spt -l` | 同期歌詞を表示 |
| `spt init` | | Client ID 設定 & Spotify 認証 |
| `spt play` | `spt p` | 再生を再開 |
| `spt stop` | `spt s` | 一時停止 |
| `spt next` | `spt n` | 次の曲へ |
| `spt back` | `spt b` | 前の曲へ |
| `spt now` | | 現在再生中の曲を表示 |
| `spt volume` | `spt v` | 音量調整 TUI を起動 |
| `spt volume [0-100]` | `spt v [0-100]` | 音量を設定 |
| `spt devices` | `spt d` | 利用可能なデバイス一覧を表示 |
| `spt settings` | | 画像サイズプリセットを変更 |

## ⚙️ 設定

`spt settings` でアルバムアートのサイズを変更できます。矢印キーで選択し、Enter で確定。

| プリセット | サイズ |
|---|---|
| small | 16×8 |
| medium | 20×10（デフォルト） |
| large | 28×14 |

## 📄 ライセンス

[Apache-2.0 LICENSE](https://github.com/T4ko0522/Spotify-CLI/blob/main/LICENSE)
