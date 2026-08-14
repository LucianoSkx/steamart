<!--
  SPDX-FileCopyrightText: 2026 SteamArt contributors
  SPDX-License-Identifier: MIT
-->

<p align="center">
  <img src="assets/icon.png" alt="SteamArt logo" width="128" height="128">
</p>

<h1 align="center">SteamArt</h1>

<p align="center">
  <strong>Metadados e arte (capa/hero/logo/ícone) para atalhos não-Steam na biblioteca do Steam</strong><br>
  <em>Usa o catálogo oficial da Steam (CDN) ou a comunidade SteamGridDB</em>
</p>

<p align="center">
  <a href="https://github.com/LucianoSkx/steamart/releases/latest">
    <img alt="Latest Release" src="https://img.shields.io/github/v/release/LucianoSkx/steamart?label=version&style=flat-square">
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
  <a href="#-português-br">🇧🇷 Português (BR)</a> •
  <a href="#-english">🇺🇸 English</a>
</p>

---

## 🇧🇷 Português (BR)

### Visão geral

O **SteamArt** reúne o que fazem o [Decky-Metadata](https://github.com/beallio/Decky-Metadata) e o [decky-steamgriddb](https://github.com/SteamGridDB/decky-steamgriddb) em um **app nativo Linux** (Go + Fyne).  

Ele corresponde seus jogos **não-Steam** ao catálogo da Steam e aplica **capas, hero, logo e ícone** — seja baixando a arte oficial do CDN da Steam, seja escolhendo imagens da comunidade no **SteamGridDB** (com filtro para arte animada).

> ✅ Não depende do Decky Loader nem do modo Big Picture  
> ✅ Escreve direto na pasta `grid` do Steam  
> ✅ Backup automático antes de sobrescrever (`grid/backup/`)  
> ✅ Interface em **Português (BR)** e **English** (lembrado entre sessões)

---

### Compatibilidade

| Plataforma | Status | Detalhes |
|------------|--------|----------|
| **Linux (qualquer distro)** | ✅ Total | AppImage portátil roda em Debian/Ubuntu, Fedora, Arch, Mint, openSUSE, Pop!_OS, Endeavour, etc. |
| **Steam (nativo)** | ✅ Total | `~/.steam/steam`, `~/.local/share/Steam` |
| **Steam (Flatpak)** | ✅ Total | `~/.var/app/com.valvesoftware.Steam/.local/share/Steam` |
| **Steam (custom)** | ✅ Via `STEAM_ROOT` | Defina `STEAM_ROOT=/caminho/para/Steam` se necessário |

---

### Funcionalidades

| Botão | Ação |
|-------|------|
| **Auto** | Auto-match pelo nome + aplica arte oficial da Steam (CDN) |
| **Buscar Steam** | Busca manual na loja + aplica arte oficial do app escolhido |
| **Buscar imagens** | Busca no SteamGridDB + navega por grid/hero/logo/ícone + aplica a imagem clicada (filtro "só animadas") |
| **Detalhes** | Modal com descrição, desenvolvedores, gêneros, compatibilidade Steam Deck |
| **Limpar imagens** | Remove toda a arte do atalho (move para `grid/backup/`) |
| **Auto-match tudo** | Aplica match + arte em lote nos atalhos pendentes |

---

### Instalação

#### 📦 Opção 1 — AppImage (recomendado, qualquer distro)

```bash
# 1. Baixe o AppImage na página de Releases
# 2. Dê permissão de execução e rode
chmod +x SteamArt-*.AppImage
./SteamArt-*.AppImage
```

> **Sem instalação, sem root, roda em qualquer distro.**  
> Baixe em: [Releases](https://github.com/LucianoSkx/steamart/releases)

#### 🛠️ Opção 2 — Script de instalação (compila do fonte)

```bash
# Instalação local (sem sudo, em ~/.local)
bash install.sh

# Instalação system-wide (precisa sudo, em /usr/local)
bash install.sh --system
```

#### ⚙️ Opção 3 — Makefile (para desenvolvedores/empacotadores)

```bash
make              # compila ./steamart
sudo make install # instala no sistema (DESTDIR suportado para empacotamento)
make appimage     # gera AppImage portátil
make test         # roda testes
make clean        # limpa build
```

---

### Pré-requisitos

| Tipo | Requisito |
|------|-----------|
| **Runtime** | Steam instalado com ≥1 atalho não-Steam |
| **Build (opcional)** | Go 1.22+ + dependências do Fyne |

**Dependências de build por distro:**

```bash
# Debian / Ubuntu / Mint / Pop!_OS
sudo apt install golang gcc libgtk-3-dev libgl1-mesa-dev libglu1-mesa-dev libx11-dev xorg-dev

# Fedora
sudo dnf install golang gcc gtk3-devel mesa-libGL-devel mesa-libGLU-devel libX11-devel libXrandr-devel libXcursor-devel libXinerama-devel libXi-devel

# Arch / Endeavour / Manjaro
sudo pacman -S go gcc gtk3 mesa libx11 libxrandr libxcursor libxinerama libxi
```

> **SteamGridDB** (opcional): crie uma API key gratuita em https://www.steamgriddb.com/api e informe na UI do app.

---

### Como usar

1. Abra o SteamArt — ele detecta automaticamente o Steam, o usuário ativo e a pasta `grid`
2. Para cada atalho não-Steam, use os botões descritos acima
3. **Auto-match tudo (N)** aplica em lote nos atalhos sem arte
4. Reinicie/abra a biblioteca do Steam para ver as artes aplicadas

---

### Build do zero

```bash
# Binário nativo (app principal)
go build -o steamart ./cmd/gui
./steamart

# Servidor web opcional (mesma funcionalidade, UI HTML)
go build -o steamart-web ./cmd/legacy-server
./steamart-web            # http://127.0.0.1:8731

# AppImage portátil
bash build-appimage.sh    # gera SteamArt-vX.Y.Z-x86_64.AppImage
```

---

### Detalhes técnicos

- Lê `shortcuts.vdf` (formato binário Valve) e usa o `appid` gravado diretamente
- Nomes de arquivo na grid:
  - `<appid>p.<ext>` — grid (capa vertical)
  - `<appid>_hero.<ext>` — hero (banner horizontal)
  - `<appid>_logo.<ext>` — logo
  - `<appid>_icon.<ext>` — ícone
  - `<appid>.<ext>` — capa horizontal (sem sufixo)
- Steam CDN: `shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/<APP_ID>/...`
- Store search: `store.steampowered.com/api/storesearch` (cc=BR, l=brazilian)
- Chave SteamGridDB salva em `<steam_config>/steamart-sgdb.json` (perms 0600)
- Logs em `<steam_config>/steamart.log`

---

### Estrutura do projeto

```
cmd/
├── gui/            # App nativo Fyne (binário principal)
├── legacy-server/  # Servidor web opcional (UI HTML/JS)
├── debug/          # Utilitários de depuração
└── genicon/        # Gera o ícone (assets/icon.png)

internal/
├── vdf/            # Parser de shortcuts.vdf
├── steam/          # Descoberta do Steam, atalhos e grid
├── match/          # Busca/auto-match e metadados da loja Steam
├── artwork/        # Download de arte (CDN + URLs) para a grid
├── sgdb/           # Cliente da API SteamGridDB
├── store/          # JSON local de matches + logs
├── delisted/       # Índice de jogos delisted da Steam
├── title/          # Normalização de títulos para matching
└── i18n/           # Traduções PT-BR / EN
```

---

### Contribuindo

1. Fork o repo
2. Crie uma branch: `git checkout -b feature/minha-feature`
3. Commit: `git commit -m "feat: minha feature"`
4. Push: `git push origin feature/minha-feature`
5. Abra um Pull Request

> Código segue `gofmt`, `go vet` e testes (`go test ./...`). O CI roda tudo automaticamente.

---

### Licença

**MIT** — veja [LICENSE](LICENSE).

---

## 🇺🇸 English

### Overview

**SteamArt** combines the functionality of [Decky-Metadata](https://github.com/beallio/Decky-Metadata) and [decky-steamgriddb](https://github.com/SteamGridDB/decky-steamgriddb) into a **native Linux app** (Go + Fyne).  

It matches your **non-Steam games** to the Steam catalog and applies **grid, hero, logo, and icon** artwork — either by downloading official art from the Steam CDN or by selecting community images from **SteamGridDB** (with animated-art filter).

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
make clean        # cleans build artifacts
```

---

### Prerequisites

| Type | Requirement |
|------|-------------|
| **Runtime** | Steam installed with ≥1 non-Steam shortcut |
| **Build (optional)** | Go 1.22+ + Fyne system dependencies |

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

# Optional web server (same features, HTML UI)
go build -o steamart-web ./cmd/legacy-server
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
├── steam/          # Steam discovery, shortcuts & grid
├── match/          # Store search/auto-match & metadata
├── artwork/        # Art download (CDN + URLs) to grid
├── sgdb/           # SteamGridDB API client
├── store/          # Local JSON matches + logs
├── delisted/       # Steam delisted games index
├── title/          # Title normalization for matching
└── i18n/           # PT-BR / EN translations
```

---

### Contributing

1. Fork the repo
2. Create a branch: `git checkout -b feature/my-feature`
3. Commit: `git commit -m "feat: my feature"`
4. Push: `git push origin feature/my-feature`
5. Open a Pull Request

> Code follows `gofmt`, `go vet` and tests (`go test ./...`). CI runs all checks automatically.

---

### License

**MIT** — see [LICENSE](LICENSE).