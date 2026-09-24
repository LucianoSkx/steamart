# Changelog

Todas as mudanças notáveis deste projeto são documentadas aqui.
Formato baseado em [Keep a Changelog](https://keepachangelog.com/) e
[Versionamento Semântico](https://semver.org/).

## [Não lançado]

### Corrigido
- `steamart-matches.json` corrompido não destrói mais as associações: o arquivo
  ruim é preservado em `.corrompido-<data>` e o app inicia vazio com aviso
- Gravações na grid agora são atômicas (tmp + rename), evitando arte truncada
  enquanto a Steam lê a pasta
- Backup da arte anterior não é mais sobrescrito: versões antigas ficam em
  `grid/backup/` com sufixo numérico (`-2`, `-3`…)
- Download do índice de jogos removidos tem timeout de 30s (antes pendurava o
  startup do servidor e o "Auto Steam" da GUI)
- Match por nome não falha mais em títulos acentuados (`Pokémon`/`Pokemon`,
  `Ragnarök`/`Ragnarok`)
- Remoção/listagem de arte não confunde atalhos por prefixo (`123` não pega
  mais a arte de `1234`) e agora cobre a capa horizontal (`<appid>.jpg`)
- `/api/match` valida se o atalho existe antes de gravar arte órfã na grid
- Parser VDF reporta erro em `int32`/string truncados em vez de devolver 0
  silenciosamente
- Corrida de dados na troca de idioma da UI (`i18n`)

### Adicionado
- Testes para `vdf`, `steam`, `store`, `grid`, `sgdb`, `official` e cobertura
  de acentos em `title`/`match` (pacotes testados: 4 → 10)
- Pacote `internal/atomicfile` (escrita atômica) e `internal/grid` (regras de
  arquivos da grid, compartilhadas por GUI e servidor legado)
- CI: `gofmt`, `staticcheck`, testes com `-race` e relatório de cobertura

## [v1.0.3] - 2026-08-18

### Corrigido
- Servidor legado (`cmd/legacy-server`, build tag `-tags legacy`) voltou a
  compilar: `match.AutoMatch` passou a receber o índice de jogos delisted
  (`[]delisted.App`), mantendo o comportamento de auto-match
- CI (`ci.yml`) agora compila e valida o `legacy-server` com `-tags legacy`,
  além do build/vet/test do app principal

## [v1.0.1] - 2026-08-14

### Corrigido
- AppImage agora inclui a versão no filename (`SteamArt-v1.0.0-x86_64.AppImage`)
- Versão injetada no binário via `-ldflags="-X main.Version=..."` e visível no diálogo "Sobre"
- Metainfo AppStream atualizado com versão no build
- URLs do GitHub apontam para `LucianoSkx/steamart`

### Adicionado
- README bilíngue (PT-BR / English) com badges, tabelas e seções completas
- CONTRIBUTING.md com guia de contribuição
- CHANGELOG.md (este arquivo)

## [v1.0.0] - 2026-08-14

Lançamento inicial.

### Funcionalidades
- App nativo Fyne (Go) para aplicar arte (grid/hero/logo/ícone) em atalhos não-Steam do Steam
- Auto-match de jogos não-Steam com o catálogo da Steam (CDN oficial)
- Busca e aplicação de imagens do SteamGridDB (com filtro de arte animada)
- Diálogo "Sobre" com informações da versão e créditos
- Interface bilíngue: Português (BR) / English
- Backup automático de arte anterior (`grid/backup/`)
- Remoção de arte de atalhos
- Detalhes do jogo (descrição, devs, gêneros, Steam Deck)
- Servidor web opcional (`cmd/legacy-server`) com UI HTML
- Descoberta automática do Steam (nativo, Flatpak e via `STEAM_ROOT`)

### Distribuição
- AppImage portátil (build automático via GitHub Actions)
- Script de instalação (`install.sh`) — local ou system-wide
- Makefile com targets: build, install, test, appimage, clean
- Metadados AppStream válidos
- Arquivo `.desktop` genérico e ícone 512×512
