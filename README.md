# SteamArt

Programa standalone (Linux desktop) que reúne o que fazem o
[Decky-Metadata](https://github.com/beallio/Decky-Metadata) e o
[decky-steamgriddb](https://github.com/SteamGridDB/decky-steamgriddb):
corresponde seus jogos não-Steam ao catálogo da Steam e aplica capas/hero/logo/ícone,
tanto via Steam (CDN) quanto via imagens da comunidade do SteamGridDB.

Não depende do Decky Loader nem do modo Big Picture — roda como um programa comum
e escreve direto na pasta `grid` do Steam.

## Funcionalidades
- **Metadados Steam**: lista os atalhos não-Steam, faz auto-match pelo nome e
  baixa a arte oficial (capa/hero/logo/ícone) do CDN da Steam para a grid.
- **SteamGridDB**: busca o jogo na base da comunidade e aplica imagens de
  grid/hero/logo/ícone escolhidas por você (com filtro de arte animada).
- **Auto-match tudo**: aplica match + arte em todos os atalhos que ainda não
  têm arte na grid, de uma só vez (preserva arte existente).
- **Remover arte**: apaga a arte (capa/hero/logo/ícone) de um atalho.
- **Detalhes**: abre um modal com descrição, desenvolvedores, gêneros e
  compatibilidade Steam Deck do app correspondente.
- **Backup automático**: antes de sobrescrever, a arte anterior é movida para
  `grid/backup/`.
- **Idiomas**: seletor **Português (BR) / English** no topo da janela; a escolha
  é lembrada entre sessões.

## Pré-requisitos
- Steam instalado e com pelo menos um atalho não-Steam.
- Go 1.22+ para compilar.
- (SteamGridDB) uma API key gratuita em https://www.steamgriddb.com/api (informe na própria UI).

## Build e execução

O programa tem duas frentes:

- **App nativo (padrão)** — janela própria em Fyne, sem navegador. É o
  binário principal.
  ```bash
  go build -o steamart ./cmd/gui
  ./steamart
  ```
- **Servidor web (opcional)** — sobe um servidor HTTP com a mesma UI em
  HTML/JS (útil pra acessar de outro dispositivo).
  ```bash
  go build -o steamart-web .
  ./steamart-web            # abre em http://127.0.0.1:8731
  ```

Para instalar no sistema:
```bash
sudo cp steamart /usr/local/bin/
```

Há também um `steamart.desktop` para aparecer no menu de aplicativos
(instale em `~/.local/share/applications/`). O app nativo não abre navegador
nem precisa de servidor.

## Como usar
1. O programa detecta sozinho o Steam (`~/.local/share/Steam`), o usuário ativo
   e a pasta `grid` (`userdata/<id>/config/grid`).
2. Cada atalho tem botões:
   - **Auto**: tenta casar o nome com a loja da Steam e aplica a arte oficial.
   - **Buscar Steam**: busca manual e aplica a arte oficial do app escolhido.
   - **SteamGridDB**: busca na comunidade, navega por grid/hero/logo/ícone e
     aplica a imagem clicada (com opção "só animadas").
   - **Detalhes**: mostra metadados do app Steam correspondente.
   - **Remover arte**: apaga a arte atual do atalho.
3. O botão **Auto-match tudo (N)** no topo aplica em lote nos atalhos pendentes.
4. A arte aplicada aparece nos cartões e na própria biblioteca do Steam após
   reiniciar/reabrir a biblioteca.

## Detalhes técnicos
- Lê `shortcuts.vdf` (binário Valve) e usa o `appid` gravado diretamente
  (não recalcula via CRC — o valor bate com o que o Steam usa).
- Arquivos na grid: `<appid>p.<ext>` (grid), `<appid>_hero.<ext>`,
  `<appid>_logo.<ext>`, `<appid>_icon.<ext>`.
- Steam CDN: `shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/<M>/...`
- Store search: `store.steampowered.com/api/storesearch` (cc=BR, l=brazilian).
- A chave do SteamGridDB é salva localmente em
  `<steam_config>/steamart-sgdb.json`.
- Logs em `<steam_config>/steamart.log`.

## Estrutura
```
internal/vdf      parser de shortcuts.vdf
internal/steam    descoberta do Steam, atalhos e grid
internal/match    busca/auto-match e metadados da loja Steam
internal/artwork  download de arte (CDN e URLs arbitrárias) para a grid
internal/sgdb     cliente da API SteamGridDB
internal/store    JSON local de matches + logs (memória e arquivo)
main.go           servidor HTTP + API + UI embarcada
web/             interface (index.html, style.css, app.js)
```
