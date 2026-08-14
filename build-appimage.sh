#!/usr/bin/env bash
#
# Gera um AppImage do SteamArt (portátil, roda em qualquer distro Linux).
# Requer: Go + dependências de build do Fyne (gcc, libgtk-3-dev, libgl, libx11, xorg-dev).
set -euo pipefail

ARCH="${ARCH:-x86_64}"
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo dev)}"
APP="SteamArt"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$HERE"

# 1. binário
echo "Compilando..."
go build -ldflags="-X main.Version=${VERSION}" -o steamart ./cmd/gui

# 2. AppDir
APPDIR="AppDir"
rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications" "$APPDIR/usr/share/icons/hicolor/512x512/apps" "$APPDIR/usr/share/metainfo"
cp steamart "$APPDIR/usr/bin/steamart"
cp steamart.desktop "$APPDIR/steamart.desktop"
cp steamart.desktop "$APPDIR/usr/share/applications/steamart.desktop"
cp assets/icon.png "$APPDIR/steamart.png"
cp assets/icon.png "$APPDIR/usr/share/icons/hicolor/512x512/apps/steamart.png"
cp assets/com.steamart.app.metainfo.xml "$APPDIR/usr/share/metainfo/com.steamart.app.metainfo.xml"

# 3. linuxdeploy + plugin go (baixa se ausente)
LD="linuxdeploy-x86_64.AppImage"
if [ ! -f "$LD" ]; then
  wget -q "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage" -O "$LD"
  chmod +x "$LD"
fi
# Evita problemas de FUSE em alguns ambientes (ex.: CI)
export APPIMAGE_EXTRACT_AND_RUN=1
./"$LD" --appdir "$APPDIR" --output appimage

# 4. nome estável
mv -f "$APP-$ARCH.AppImage" "$APP-$VERSION-$ARCH.AppImage" 2>/dev/null || true
echo "Pronto: $APP-$VERSION-$ARCH.AppImage"
