#!/usr/bin/env bash
#
# Instala o SteamArt a partir do código-fonte.
#   bash install.sh          → instala em ~/.local (sem root)
#   bash install.sh --system → instala em /usr/local (precisa de sudo)
set -euo pipefail

SYSTEM=0
[ "${1:-}" = "--system" ] && SYSTEM=1

if [ "$SYSTEM" -eq 1 ]; then
  BIN_DIR="/usr/local/bin"
  APP_DIR="/usr/local/share/applications"
  ICON_DIR="/usr/local/share/icons/hicolor/512x512/apps"
  SUDO="sudo"
else
  BIN_DIR="$HOME/.local/bin"
  APP_DIR="$HOME/.local/share/applications"
  ICON_DIR="$HOME/.local/share/icons/hicolor/512x512/apps"
  SUDO=""
fi

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$HERE"

echo "Compilando SteamArt..."
go build -o steamart ./cmd/gui

META_DIR="$ICON_DIR/../../../share/metainfo"
$SUDO install -d "$BIN_DIR" "$APP_DIR" "$ICON_DIR" "$META_DIR"
$SUDO install -m755 steamart "$BIN_DIR/steamart"
$SUDO install -m644 steamart.desktop "$APP_DIR/steamart.desktop"
$SUDO install -m644 assets/icon.png "$ICON_DIR/steamart.png"
$SUDO install -m644 assets/com.steamart.app.metainfo.xml "$META_DIR/com.steamart.app.metainfo.xml"

if command -v update-desktop-database >/dev/null 2>&1; then
  $SUDO update-desktop-database "$APP_DIR" || true
fi

echo "SteamArt instalado. Rode 'steamart' no terminal ou abra pelo menu de aplicativos."
