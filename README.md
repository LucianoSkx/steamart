# SteamArt

Programa Linux (desktop) que reúne o que fazem o
[Decky-Metadata](https://github.com/beallio/Decky-Metadata) e o
[decky-steamgriddb](https://github.com/SteamGridDB/decky-steamgriddb):
corresponde seus jogos não-Steam ao catálogo da Steam e aplica capas/hero/logo/ícone,
tanto via Steam (CDN) quanto via imagens da comunidade do SteamGridDB.

Não depende do Decky Loader nem do modo Big Picture — roda como um programa comum
e escreve direto na pasta `grid` do Steam.

## Compatibilidade

- **Distros**: qualquer distribuição Linux desktop (Debian/Ubuntu, Fedora, Arch,
  Mint, openSUSE, Pop!_OS, Endeavour, etc.). O AppImage roda em todas sem instalar
  nada além de um ambiente gráfico comum.
- **Steam**: detecta automaticamente as instalações em `~/.steam/steam`,
  `~/.local/share/Steam` e **Steam via Flatpak**
  (`~/.var/app/com.valvesoftware.Steam/...`). Se nada disso servir, defina a
  variável de ambiente `STEAM_ROOT=/caminho/para/Steam`.

## Funcionalidades

- **Metadados Steam**: lista os atalhos não-Steam, faz auto-match pelo nome e
  baixa a arte oficial (capa/hero/logo/ícone) do CDN da Steam para a grid.
- **SteamGridDB**: busca o jogo na base da comunidade e aplica imagens de
  grid/hero/logo/ícone escolhidas por você (com filtro de arte animada).
- **Auto-match tudo**: aplica match + arte em todos os atalhos pendentes de uma vez.
- **Remover arte**: apaga a arte (capa/hero/logo/ícone) de um atalho.
- **Detalhes**: modal com descrição, desenvolvedores, gêneros e compatibilidade
  Steam Deck do app correspondente.
- **Backup automático**: a arte anterior é movida para `grid/backup/` antes de
  sobrescrever.
- **Idiomas**: seletor Português (BR) / English no topo da janela (lembrado entre
  sessões).

## Instalação

### Opção 1 — AppImage (recomendado, qualquer distro)

Baixe o `SteamArt-*.AppImage` na página de
[Releases](https://github.com/SEU_USUARIO/steamart/releases), dê permissão de
execução e rode:

```bash
chmod +x SteamArt-*.AppImage
./SteamArt-*.AppImage
```

### Opção 2 — Script (compila do código, sem root)

```bash
bash install.sh            # instala em ~/.local
# ou, para todo o sistema:
bash install.sh --system   # precisa de sudo
```

### Opção 3 — Makefile

```bash
make                      # compila ./steamart
sudo make install         # instala no sistema
```

## Pré-requisitos

- Steam instalado com pelo menos um atalho não-Steam.
- (Build do código) Go 1.22+ e as dependências de sistema do Fyne:
  - **Debian/Ubuntu**: `sudo apt install golang gcc libgtk-3-dev libgl1-mesa-dev libglu1-mesa-dev libx11-dev xorg-dev`
  - **Fedora**: `sudo dnf install golang gcc gtk3-devel mesa-libGL-devel mesa-libGLU-devel libX11-devel libXrandr-devel libXcursor-devel libXinerama-devel libXi-devel`
  - **Arch**: `sudo pacman -S go gcc gtk3 mesa libx11 libxrandr libxcursor libxinerama libxi`
- (SteamGridDB) uma API key gratuita em https://www.steamgriddb.com/api (informe na própria UI).

## Como usar

1. O programa detecta sozinho o Steam, o usuário ativo e a pasta `grid`.
2. Cada atalho tem botões:
   - **Auto**: casa o nome com a loja da Steam e aplica a arte oficial.
   - **Buscar Steam**: busca manual e aplica a arte oficial do app escolhido.
   - **SteamGridDB**: busca na comunidade e aplica a imagem clicada (com
     opção "só animadas").
   - **Detalhes**: mostra metadados do app Steam correspondente.
   - **Remover arte**: apaga a arte atual do atalho.
3. **Auto-match tudo (N)** aplica em lote nos atalhos pendentes.
4. A arte aparece na biblioteca do Steam após reiniciar/reabrir a biblioteca.

## Build do zero

```bash
# binário simples
go build -o steamart ./cmd/gui
./steamart

# servidor web opcional (mesma ideia, UI em HTML)
go build -o steamart-web ./cmd/legacy-server
./steamart-web            # http://127.0.0.1:8731

# AppImage portátil
bash build-appimage.sh
```

## Detalhes técnicos

- Lê `shortcuts.vdf` (binário Valve) e usa o `appid` gravado diretamente.
- Arquivos na grid: `<appid>p.<ext>` (grid), `<appid>_hero.<ext>`,
  `<appid>_logo.<ext>`, `<appid>_icon.<ext>`.
- Steam CDN: `shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/<M>/...`
- Store search: `store.steampowered.com/api/storesearch` (cc=BR, l=brazilian).
- A chave do SteamGridDB é salva em `<steam_config>/steamart-sgdb.json`.
- Logs em `<steam_config>/steamart.log`.

## Estrutura

```
cmd/gui/          app nativo em Fyne (binário principal)
cmd/legacy-server/ servidor web opcional (mesma funcionalidade, UI HTML)
cmd/debug/         utilitários de depuração
cmd/genicon/       gera o ícone (assets/icon.png)
internal/vdf/      parser de shortcuts.vdf
internal/steam/    descoberta do Steam, atalhos e grid
internal/match/    busca/auto-match e metadados da loja Steam
internal/artwork/  download de arte (CDN e URLs) para a grid
internal/sgdb/     cliente da API SteamGridDB
internal/store/    JSON local de matches + logs
internal/delisted/ índice de jogos delisted da Steam
internal/title/    normalização de títulos para matching
internal/i18n/     traduções PT-BR / EN
```

## Licença

MIT — veja [LICENSE](LICENSE).
