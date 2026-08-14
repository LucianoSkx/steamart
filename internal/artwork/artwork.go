package artwork

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const cdn = "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/%d/%s"

var client = &http.Client{Timeout: 60 * time.Second}

// targets mapeia o sufixo de arquivo na grid para o asset no CDN.
var targets = []struct {
	suffix   string
	asset    string
	fallback string
}{
	{"p", "library_600x900.jpg", ""},
	{"_hero", "library_hero.jpg", ""},
	{"_logo", "logo.png", ""},
	{"", "header.jpg", ""},
	{"_icon", "icon.png", ""},
}

// Result descreve o que foi gravado para um atalho.
type Result struct {
	Files map[string]string `json:"files"` // sufixo -> caminho gravado
}

// Download baixa as artes do app Steam (steamAppID) e grava na grid do
// atalho (shortcutAppID). Retorna os arquivos escritos.
func Download(shortcutAppID uint32, steamAppID int, gridDir string) (*Result, error) {
	if err := os.MkdirAll(gridDir, 0o755); err != nil {
		return nil, err
	}
	res := &Result{Files: map[string]string{}}
	for _, t := range targets {
		url := fmt.Sprintf(cdn, steamAppID, t.asset)
		ext := filepath.Ext(t.asset)
		data, ctype, err := fetch(url)
		if err != nil || len(data) == 0 {
			if t.fallback != "" {
				url = fmt.Sprintf(cdn, steamAppID, t.fallback)
				data, _, err = fetch(url)
			}
			if err != nil || len(data) == 0 {
				continue
			}
			ext = filepath.Ext(t.fallback)
		}
		_ = ctype
		dst := filepath.Join(gridDir, fmt.Sprintf("%d%s%s", shortcutAppID, t.suffix, ext))
		backupExisting(gridDir, filepath.Base(dst))
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return nil, err
		}
		res.Files[t.suffix] = dst
	}
	return res, nil
}

func fetch(url string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "steamart/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	return data, resp.Header.Get("Content-Type"), err
}

// Suffix retorna o sufixo de arquivo na grid para cada tipo de arte.
func Suffix(asset string) string {
	switch asset {
	case "hero":
		return "_hero"
	case "logo":
		return "_logo"
	case "icon":
		return "_icon"
	case "capsule":
		return "" // capa horizontal (landscape) da Steam: {appid}.ext, sem sufixo
	default: // grid
		return "p"
	}
}

// SaveURL baixa uma URL arbitrária e a grava na grid do atalho.
func SaveURL(url string, gridDir string, appid uint32, asset string) (string, error) {
	if err := os.MkdirAll(gridDir, 0o755); err != nil {
		return "", err
	}
	data, _, err := fetch(url)
	if err != nil || len(data) == 0 {
		return "", fmt.Errorf("falha ao baixar %s: %w", url, err)
	}
	ext := extFromURL(url)
	dst := filepath.Join(gridDir, fmt.Sprintf("%d%s%s", appid, Suffix(asset), ext))
	backupExisting(gridDir, filepath.Base(dst))
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return "", err
	}
	return dst, nil
}

// backupExisting move um arquivo já existente para grid/backup/ antes de
// sobrescrevê-lo, preservando a arte anterior.
func backupExisting(gridDir, name string) {
	src := filepath.Join(gridDir, name)
	if _, err := os.Stat(src); err != nil {
		return
	}
	bk := filepath.Join(gridDir, "backup")
	_ = os.MkdirAll(bk, 0o755)
	_ = os.Rename(src, filepath.Join(bk, name))
}

func extFromURL(url string) string {
	for i := len(url) - 1; i >= 0 && url[i] != '/'; i-- {
		if url[i] == '.' {
			e := strings.ToLower(url[i:])
			if e == ".jpg" || e == ".jpeg" || e == ".png" || e == ".webp" || e == ".gif" {
				if e == ".jpeg" {
					return ".jpg"
				}
				return e
			}
		}
	}
	return ".jpg"
}
