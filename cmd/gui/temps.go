package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// downloadTemp baixa um recurso remoto para um arquivo temporário.
func downloadTemp(url string) (string, error) {
	data, err := fetchRemote(url)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "smg-*.img")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// cleanupStaleTemps remove temporários smg-*.img com mais de 1 hora,
// resíduos de galerias anteriores que não foram limpos.
func cleanupStaleTemps() {
	dir := os.TempDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-time.Hour)
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "smg-") {
			continue
		}
		if info, err := e.Info(); err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
