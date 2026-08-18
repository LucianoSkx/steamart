# SteamArt — Makefile
#
# Alvos principais:
#   make            compila o binário ./steamart
#   make install    instala no sistema (precisa de sudo/DESTDIR)
#   make test       roda os testes
#   make appimage   gera o AppImage via build-appimage.sh
#   make legacy     compila o servidor HTTP legado (build tag -tags legacy)
#   make check      build + vet (gui e legado) + testes
#   make clean      remove o binário

BINARY  := steamart
PKG     := ./cmd/gui
PREFIX  ?= /usr/local
DATADIR := $(PREFIX)/share
DESTDIR ?=

.PHONY: all build legacy install uninstall test fmt vet vet-legacy check appimage clean

all: build

build:
	go build -o $(BINARY) $(PKG)

legacy:
	go build -tags legacy -o /dev/null ./cmd/legacy-server

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

vet-legacy:
	go vet -tags legacy ./cmd/legacy-server/...

check: build legacy vet vet-legacy test

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
