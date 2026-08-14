# SteamArt — Makefile
#
# Alvos principais:
#   make            compila o binário ./steamart
#   make install    instala no sistema (precisa de sudo/DESTDIR)
#   make test       roda os testes
#   make appimage   gera o AppImage via build-appimage.sh
#   make clean      remove o binário

BINARY  := steamart
PKG     := ./cmd/gui
PREFIX  ?= /usr/local
DATADIR := $(PREFIX)/share
DESTDIR ?=

.PHONY: all build install uninstall test fmt vet appimage clean

all: build

build:
	go build -o $(BINARY) $(PKG)

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

install: build
	install -Dm755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	install -Dm644 steamart.desktop $(DESTDIR)$(DATADIR)/applications/steamart.desktop
	install -Dm644 assets/icon.png $(DESTDIR)$(DATADIR)/icons/hicolor/512x512/apps/steamart.png
	install -Dm644 assets/com.steamart.app.metainfo.xml $(DESTDIR)$(DATADIR)/metainfo/com.steamart.app.metainfo.xml
	-update-desktop-database $(DESTDIR)$(DATADIR)/applications 2>/dev/null || true

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	rm -f $(DESTDIR)$(DATADIR)/applications/steamart.desktop
	rm -f $(DESTDIR)$(DATADIR)/icons/hicolor/512x512/apps/steamart.png

appimage:
	bash build-appimage.sh

clean:
	rm -f $(BINARY)
