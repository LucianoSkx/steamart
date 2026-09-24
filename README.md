<!--
  SPDX-FileCopyrightText: 2026 SteamArt contributors
  SPDX-License-Identifier: MIT
-->

<p align="center">
  <img src="assets/icon.png" alt="SteamArt logo" width="128" height="128">
</p>

<h1 align="center">SteamArt</h1>

  <p align="center">
  <strong>Aplica metadados e arte (capa/hero/logo/ícone) para atalhos não-Steam na biblioteca do Steam</strong><br>
  <em>Usa o catálogo oficial da Steam (CDN) ou a comunidade SteamGridDB — app nativo portátil, qualquer distro</em>
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
  <a href="#-português-br">🇧🇷 Português (BR)</a> •
  <a href="README.en.md">🇺🇸 English</a>
</p>

---

## 🇧🇷 Português (BR)

### Visão geral

O **SteamArt** é um app **nativo Linux** (Go + Fyne) — sem browser, sem roda,
sem dependência de interface web.  

Ele substitui o fluxo "abrir navegador → copiar → colar na grid" por: basta
rodar o programa que ele lê sua `shortcuts.vdf`, faz o matching e aplica a arte
diretamente na pasta `grid` que o Steam lê. Ideal para quem usa emuladores,
jogos da GOG, ou qualquer jogo adicionado manualmente na biblioteca.

**Como funciona:** detecta a Steam (nativa, Flatpak ou Snap), lista seus atalhos não-Steam,
casa cada jogo com o app da loja via nome + heurísticas de título, e baixa a arte
correta (grid/hero/logo/icon). Você escolhe entre a arte **oficial da Steam**
(CDN, alta qualidade) ou a **comunidade SteamGridDB** (mais opções, incluindo
arte animada/webp). Tudo com backup automático antes de sobrescrever.

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
| **Steam (Snap)** | ✅ Total | `~/snap/steam/common/.local/share/Steam` |
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

**Pré-requisito de runtime:** o AppImage precisa que seu sistema tenha uma
interface gráfica com **libGL** e **X11/Wayland** (presentes na maioria das distros
desktop). Em servidores headless, instale `libgl1` e `libgtk-3-0`.

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
make check        # build + vet (GUI e legado) + testes
make legacy       # valida o servidor legado (-tags legacy)
make clean        # limpa build
```

---

### Pré-requisitos

| Tipo | Requisito |
|------|-----------|
| **Runtime** | Steam instalado com ≥1 atalho não-Steam |
| **Build (opcional)** | Go 1.26+ + dependências do Fyne |

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

# Servidor web opcional (mesma funcionalidade, UI HTML) — build tag "legacy"
go build -tags legacy -o steamart-web ./cmd/legacy-server
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
- Escritas na grid e no store são atômicas (tmp + rename); arte anterior fica
  em `grid/backup/`, e um `steamart-matches.json` corrompido é preservado em
  `<arquivo>.corrompido-<data>` em vez de sobrescrito
- Requisições com falha transitória (429/5xx) são retentadas com backoff (`internal/httpx`)

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
├── steam/          # Descoberta do Steam (nativo/Flatpak/Snap), atalhos e grid
├── match/          # Busca/auto-match e metadados da loja Steam
├── artwork/        # Download de arte (CDN + URLs) para a grid
├── sgdb/           # Cliente da API SteamGridDB
├── store/          # JSON local de matches + logs
├── delisted/       # Índice de jogos delisted da Steam
├── official/       # Assets oficiais do CDN da Steam (grid/hero/logo/icon)
├── grid/           # Regras de nomes/lock/bkp dos arquivos da grid
├── atomicfile/     # Escrita atômica (tmp + rename)
├── httpx/          # Requisições HTTP com retry/backoff e contexto
├── title/          # Normalização de títulos para matching
├── i18n/           # Traduções PT-BR / EN
└── icon/           # Ícone embutido (PNG)
```

---

### Contribuindo

1. Fork o repo
2. Crie uma branch: `git checkout -b feature/minha-feature`
3. Commit: `git commit -m "feat: minha feature"`
4. Push: `git push origin feature/minha-feature`
5. Abra um Pull Request

> Rode `make check` (build GUI + legado + vet + testes). O CI também roda `gofmt`, `staticcheck` e `go test -race`.

---

### Licença

**MIT** — veja [LICENSE](LICENSE).
