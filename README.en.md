<!--
  SPDX-FileCopyrightText: 2026 SteamArt contributors
  SPDX-License-Identifier: MIT
-->

<p align="center">
  <img src="assets/icon.png" alt="SteamArt logo" width="128" height="128">
</p>

<h1 align="center">SteamArt</h1>

  <p align="center">
  <strong>Applies metadata and artwork (cover/hero/logo/icon) to non-Steam shortcuts in your Steam library</strong><br>
  <em>Uses the official Steam catalog (CDN) or the SteamGridDB community — portable native app, any distro</em>
</p>

<p align="center">
  <a href="https://github.com/LucianoSkx/steamart/releases/latest">
    <img alt="Latest Release" src="https://img.shields.io/github/v/release/LucianoSkx/steamart?label=version&style=flat-square">
  </a>
  <a href="https://github.com/LucianoSkx/steamart/actions/workflows/ci.yml">
    <img alt="CI Status" src="https://img.shields.io/github/actions/workflow/status/LucianoSkx/steamart/ci.yml?style=flat-square">
  </a>
  <a href="https://github.com/LucianoSkx/steamart/actions/workflows/release.yml">
    <img alt="Build Status" src="https://img.shields.io/github/actions/workflow/status/LucianoSkx/steamart/release.yml?style=flat-square">
  </a>
  <a href="LICENSE">
    <img alt="License" src="https://img.shields.io/github/license/LucianoSkx/steamart?style=flat-square">
  </a>
  <a href="https://goreportcard.com/report/github.com/LucianoSkx/steamart">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/LucianoSkx/steamart?style=flat-square">
  </a>
  <a href="https://pkg.go.dev/github.com/LucianoSkx/steamart">
    <img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/LucianoSkx/steamart?style=flat-square">
  </a>
</p>

<p align="center">
  <a href="README.md">🇧🇷 Português (BR)</a> •
  <a href="#-english">🇺🇸 English</a>
</p>

---

## 🇺🇸 English

### Overview

**SteamArt** is a **native Linux app** (Go + Fyne) — no browser, no web server,
no external dependencies at runtime beyond a standard desktop environment.  

It replaces the "open browser → copy → paste into grid" workflow with a simple
GUI: just run the program, and it reads your `shortcuts.vdf`, matches each
non-Steam game to its Steam store listing, and applies the correct artwork
(grid/hero/logo/icon) directly to your `grid` folder. Perfect for emulator
games, GOG imports, or anything you've added manually to your Steam library.

**How it works:** detects Steam (native, Flatpak or Snap), lists your non-Steam
shortcuts, matches each game to the store via title heuristics, and downloads
the right art — either **official Steam CDN** assets (highest quality) or
**SteamGridDB community** images (more options, including animated/webp).
Everything is backed up first to `grid/backup/` before overwriting.

> ✅ No Decky Loader or Big Picture mode required  
> ✅ Writes directly to Steam's `grid` folder  
> ✅ Automatic backup before overwriting (`grid/backup/`)  
> ✅ UI in **Portuguese (BR)** and **English** (persisted across sessions)

---

### Compatibility

| Platform | Status | Details |
|----------|--------|---------|
| **Linux (any distro)** | ✅ Full | Portable AppImage runs on Debian/Ubuntu, Fedora, Arch, Mint, openSUSE, Pop!_OS, Endeavour, etc. |
| **Steam (native)** | ✅ Full | `~/.steam/steam`, `~/.local/share/Steam` |
| **Steam (Flatpak)** | ✅ Full | `~/.var/app/com.valvesoftware.Steam/.local/share/Steam` |
| **Steam (Snap)** | ✅ Full | `~/snap/steam/common/.local/share/Steam` |
| **Steam (custom path)** | ✅ Via `STEAM_ROOT` | Set `STEAM_ROOT=/path/to/Steam` if needed |

---

### Features

| Button | Action |
|--------|--------|
| **Auto** | Auto-match by name + applies official Steam art (CDN) |
| **Search Steam** | Manual store search + applies official art of chosen app |
| **Search images** | Search SteamGridDB + browse grid/hero/logo/icon + apply clicked image (animated-only filter) |
| **Details** | Modal with description, developers, genres, Steam Deck compatibility |
| **Clear images** | Removes all art from shortcut (moves to `grid/backup/`) |
| **Auto-match all** | Batch applies match + art to all shortcuts missing art |

---

### Installation

#### 📦 Option 1 — AppImage (recommended, any distro)

```bash
# 1. Download the AppImage from Releases
# 2. Make executable and run
chmod +x SteamArt-*.AppImage
./SteamArt-*.AppImage
```

> **No install, no root, runs on any distro.**  
> Download at: [Releases](https://github.com/LucianoSkx/steamart/releases)

#### 🛠️ Option 2 — Install script (builds from source)

```bash
# Local install (no sudo, into ~/.local)
bash install.sh

# System-wide install (needs sudo, into /usr/local)
bash install.sh --system
```

#### ⚙️ Option 3 — Makefile (for developers/packagers)

```bash
make              # builds ./steamart
sudo make install # installs system-wide (DESTDIR supported for packaging)
make appimage     # builds portable AppImage
make test         # runs tests
make check        # build + vet (GUI and legacy) + tests
make legacy       # builds the legacy server (-tags legacy)
make clean        # cleans build artifacts
```

---

### Prerequisites

| Type | Requirement |
|------|-------------|
| **Runtime** | Steam installed with ≥1 non-Steam shortcut |
| **Build (optional)** | Go 1.26+ + Fyne system dependencies |

**Build dependencies by distro:**

```bash
# Debian / Ubuntu / Mint / Pop!_OS
sudo apt install golang gcc libgtk-3-dev libgl1-mesa-dev libglu1-mesa-dev libx11-dev xorg-dev

# Fedora
sudo dnf install golang gcc gtk3-devel mesa-libGL-devel mesa-libGLU-devel libX11-devel libXrandr-devel libXcursor-devel libXinerama-devel libXi-devel

# Arch / Endeavour / Manjaro
sudo pacman -S go gcc gtk3 mesa libx11 libxrandr libxcursor libxinerama libxi
```

> **SteamGridDB** (optional): get a free API key at https://www.steamgriddb.com/api and enter it in the app UI.

---

### Usage

1. Launch SteamArt — it auto-detects Steam, the active user, and the `grid` folder
2. For each non-Steam shortcut, use the buttons described above
3. **Auto-match all (N)** batch-processes shortcuts missing art
4. Restart/reopen Steam library to see applied artwork

---

### Build from source

```bash
# Native binary (main app)
go build -o steamart ./cmd/gui
./steamart

# Optional web server (same features, HTML UI) — requires the "legacy" build tag
go build -tags legacy -o steamart-web ./cmd/legacy-server
./steamart-web            # http://127.0.0.1:8731

# Portable AppImage
bash build-appimage.sh    # produces SteamArt-vX.Y.Z-x86_64.AppImage
```

---

### Technical details

- Parses `shortcuts.vdf` (Valve binary format) and uses the stored `appid` directly
- Grid filenames:
  - `<appid>p.<ext>` — grid (vertical cover)
  - `<appid>_hero.<ext>` — hero (horizontal banner)
  - `<appid>_logo.<ext>` — logo
  - `<appid>_icon.<ext>` — icon
  - `<appid>.<ext>` — horizontal cover (no suffix)
- Steam CDN: `shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/<APP_ID>/...`
- Store search: `store.steampowered.com/api/storesearch` (cc=BR, l=brazilian)
- SteamGridDB key stored at `<steam_config>/steamart-sgdb.json` (0600 perms)
- Logs at `<steam_config>/steamart.log`
- Grid and store writes are atomic (tmp + rename); previous art is kept in
  `grid/backup/`, and a corrupted `steamart-matches.json` is preserved as
  `<file>.corrupt-<date>` instead of being overwritten
- Transient request failures (429/5xx) are retried with backoff (`internal/httpx`)

---

### Project structure

```
cmd/
├── gui/            # Native Fyne app (main binary)
├── legacy-server/  # Optional web server (HTML/JS UI)
├── debug/          # Debug utilities
└── genicon/        # Generates icon (assets/icon.png)

internal/
├── vdf/            # shortcuts.vdf parser
├── steam/          # Steam discovery (native/Flatpak/Snap), shortcuts & grid
├── match/          # Store search/auto-match & metadata
├── artwork/        # Art download (CDN + URLs) to grid
├── sgdb/           # SteamGridDB API client
├── store/          # Local JSON matches + logs
├── delisted/       # Steam delisted games index
├── official/       # Official Steam CDN assets (grid/hero/logo/icon)
├── grid/           # Grid file naming/lock/backup rules
├── atomicfile/     # Atomic writes (tmp + rename)
├── httpx/          # HTTP requests with retry/backoff and context
├── title/          # Title normalization for matching
├── i18n/           # PT-BR / EN translations
└── icon/           # Embedded icon (PNG)
```

---

### Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for the full guide. In short:

1. Fork → branch → code → test (`make check`) → PR
2. Follow `gofmt`, `go vet` and keep tests green (CI also runs `staticcheck` and `go test -race`)
3. Comments in Portuguese (pt-BR)
4. Releases are automatic via tags (`git tag vX.Y.Z && git push --tags`)

See [CHANGELOG.md](CHANGELOG.md) for the version history.

---

### License

**MIT** — see [LICENSE](LICENSE).
