//go:build legacy

// Servidor HTTP antigo (antes da GUI Fyne). Compile apenas com
// -tags legacy; fora do build padrão.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"steamart/internal/artwork"
	"steamart/internal/delisted"
	"steamart/internal/match"
	"steamart/internal/sgdb"
	"steamart/internal/steam"
	"steamart/internal/store"
)

//go:embed web/*
var webFS embed.FS

var port = 8731

var (
	steamClient   *steam.Steam
	matches       *store.Store
	logger        = &store.Logger{}
	sgdbKey       string
	delistedIndex *delisted.Index
)

type shortcutView struct {
	steam.Shortcut
	Artwork map[string]string `json:"artwork"`
	Match   *store.Match      `json:"match"`
}

func main() {
	flag.IntVar(&port, "port", 8731, "porta do servidor HTTP")
	flag.Parse()

	s, err := steam.Discover()
	if err != nil {
		log.Fatalf("não encontrei a Steam: %v", err)
	}
	steamClient = s
	matches, err = store.Open(s.Config)
	if err != nil {
		log.Fatalf("não abri o store: %v", err)
	}
	if idx := delisted.Ensure(nil, filepath.Join(s.Config, "delisted_index.json"), false); idx != nil {
		delistedIndex = idx
	}
	_ = logger.SetFile(filepath.Join(s.Config, "steamart.log"))
	if b, rerr := os.ReadFile(filepath.Join(s.Config, "steamart-sgdb.json")); rerr == nil {
		var k struct {
			Key string `json:"key"`
		}
		if json.Unmarshal(b, &k) == nil {
			sgdbKey = k.Key
		}
	}
	logger.Add(fmt.Sprintf("Steam em %s (usuário %s)", s.Root, s.UserID))

	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/shortcuts", handleShortcuts)
	http.HandleFunc("/api/search", handleSearch)
	http.HandleFunc("/api/meta", handleMeta)
	http.HandleFunc("/api/automatch", handleAutoMatch)
	http.HandleFunc("/api/automatch-all", handleAutoMatchAll)
	http.HandleFunc("/api/match", handleApply)
	http.HandleFunc("/api/remove", handleRemove)

	http.HandleFunc("/api/sgdb/status", handleSGDBStatus)
	http.HandleFunc("/api/sgdb/key", handleSGDBKey)
	http.HandleFunc("/api/sgdb/search", handleSGDBSearch)
	http.HandleFunc("/api/sgdb/images", handleSGDBImages)
	http.HandleFunc("/api/sgdb/apply", handleSGDBApply)

	http.HandleFunc("/grid/", handleGridFile)

	http.HandleFunc("/api/art", handleArt)

	http.HandleFunc("/api/logs", handleLogs)

	subFS, _ := fs.Sub(webFS, "web")
	http.Handle("/", http.FileServer(http.FS(subFS)))

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if len(os.Args) > 1 && os.Args[1] == "-open" {
		go openBrowser("http://" + addr)
	}
	logger.Add(fmt.Sprintf("servidor em http://%s", addr))
	log.Printf("steamart rodando em http://%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"steam_root": steamClient.Root,
		"user_id":    steamClient.UserID,
		"grid":       steamClient.Grid,
	})
}

func handleShortcuts(w http.ResponseWriter, r *http.Request) {
	list, err := steamClient.Shortcuts()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out := make([]shortcutView, 0, len(list))
	for _, sc := range list {
		out = append(out, shortcutView{
			Shortcut: sc,
			Artwork:  steamClient.HasArtwork(sc.AppID),
			Match:    matches.Get(sc.AppID),
		})
	}
	writeJSON(w, out)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q vazio", 400)
		return
	}
	res, err := match.Search(q)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, res)
}

func handleMeta(w http.ResponseWriter, r *http.Request) {
	appid := atoi(r.URL.Query().Get("appid"))
	if appid == 0 {
		http.Error(w, "appid inválido", 400)
		return
	}
	m, err := match.GetMeta(appid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, m)
}

func handleAutoMatch(w http.ResponseWriter, r *http.Request) {
	appid := uint32(atoi(r.URL.Query().Get("appid")))
	sc, err := findShortcut(appid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	res, err := match.AutoMatch(sc.AppName, delistedApps())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if res == nil {
		writeJSON(w, map[string]any{"found": false})
		return
	}
	writeJSON(w, map[string]any{"found": true, "result": res})
}

func handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST necessário", 405)
		return
	}
	var body struct {
		ShortcutAppID uint32 `json:"shortcut_appid"`
		SteamAppID    int    `json:"steam_appid"`
		Name          string `json:"name"`
		Pinned        bool   `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if body.SteamAppID <= 0 {
		http.Error(w, "steam_appid inválido", 400)
		return
	}
	res, err := artwork.Download(body.ShortcutAppID, body.SteamAppID, steamClient.Grid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	m := &store.Match{
		ShortcutAppID: body.ShortcutAppID,
		SteamAppID:    body.SteamAppID,
		Name:          body.Name,
		Pinned:        body.Pinned,
	}
	matches.Set(m)
	if err := matches.Save(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	logger.Add(fmt.Sprintf("arte aplicada: %s -> app %d (%d arquivos)", body.Name, body.SteamAppID, len(res.Files)))
	writeJSON(w, res)
}

// ------------------------------------------------------------------
// SteamGridDB
// ------------------------------------------------------------------
func handleSGDBStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"configured": sgdbKey != ""})
}

func handleSGDBKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST necessário", 405)
		return
	}
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	sgdbKey = strings.TrimSpace(body.Key)
	if err := steamClient.SaveJSON("steamart-sgdb.json", map[string]string{"key": sgdbKey}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	logger.Add("SteamGridDB API key salva")
	writeJSON(w, map[string]any{"configured": sgdbKey != ""})
}

func handleSGDBSearch(w http.ResponseWriter, r *http.Request) {
	if sgdbKey == "" {
		http.Error(w, "API key não configurada", 400)
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q vazio", 400)
		return
	}
	games, err := sgdb.New(sgdbKey).Search(q)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, games)
}

func handleSGDBImages(w http.ResponseWriter, r *http.Request) {
	if sgdbKey == "" {
		http.Error(w, "API key não configurada", 400)
		return
	}
	gameID := atoi(r.URL.Query().Get("game_id"))
	asset := r.URL.Query().Get("asset")
	if gameID == 0 || asset == "" {
		http.Error(w, "game_id ou asset inválido", 400)
		return
	}
	valid := false
	for _, a := range append(sgdb.AssetTypes, "capsule") {
		if a == asset {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "asset deve ser grid/hero/logo/icon/capsule", 400)
		return
	}
	animated := r.URL.Query().Get("animated") == "true"
	dims := r.URL.Query().Get("dimensions")
	apiAsset := asset
	if asset == "capsule" {
		apiAsset = "grid"
		if dims == "" {
			dims = "920x430"
		}
	}
	images, err := sgdb.New(sgdbKey).Images(gameID, apiAsset, dims, animated)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, images)
}

func handleGridFile(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(strings.TrimPrefix(r.URL.Path, "/grid/"))
	if name == "" || name == "." || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}
	data, err := os.ReadFile(filepath.Join(steamClient.Grid, name))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ct := "image/png"
	switch {
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"):
		ct = "image/jpeg"
	case strings.HasSuffix(name, ".webp"):
		ct = "image/webp"
	case strings.HasSuffix(name, ".gif"):
		ct = "image/gif"
	}
	w.Header().Set("Content-Type", ct)
	w.Write(data)
}

func handleArt(w http.ResponseWriter, r *http.Request) {
	appid := atoi(r.URL.Query().Get("appid"))
	if appid == 0 {
		http.Error(w, "appid inválido", 400)
		return
	}
	prefix := fmt.Sprintf("%d", appid)
	entries, err := os.ReadDir(steamClient.Grid)
	if err != nil {
		writeJSON(w, []any{})
		return
	}
	var out []map[string]string
	for _, e := range entries {
		n := e.Name()
		if !strings.HasPrefix(n, prefix) {
			continue
		}
		typ := "grid"
		switch {
		case strings.Contains(n, "_hero."):
			typ = "hero"
		case strings.Contains(n, "_logo."):
			typ = "logo"
		case strings.Contains(n, "_icon."):
			typ = "icon"
		}
		out = append(out, map[string]string{"type": typ, "url": "/grid/" + n, "file": n})
	}
	writeJSON(w, out)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, logger.All())
}

func handleSGDBApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST necessário", 405)
		return
	}
	var body struct {
		ShortcutAppID uint32 `json:"shortcut_appid"`
		URL           string `json:"url"`
		Asset         string `json:"asset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if _, err := findShortcut(body.ShortcutAppID); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	dst, err := artwork.SaveURL(body.URL, steamClient.Grid, body.ShortcutAppID, body.Asset)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	logger.Add(fmt.Sprintf("SteamGridDB: arte %s aplicada ao atalho %d", body.Asset, body.ShortcutAppID))
	writeJSON(w, map[string]string{"file": dst})
}

// handleAutoMatchAll aplica match + arte em todos os atalhos sem arte na grid,
// preservando qualquer arte já existente (manual ou anterior).
func handleAutoMatchAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST necessário", 405)
		return
	}
	list, err := steamClient.Shortcuts()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	summary := map[string]any{
		"total":    len(list),
		"applied":  0,
		"skipped":  0,
		"no_match": 0,
		"errors":   []map[string]string{},
	}
	addErr := func(name, msg string) {
		summary["errors"] = append(summary["errors"].([]map[string]string), map[string]string{"name": name, "error": msg})
	}
	for _, sc := range list {
		if len(steamClient.HasArtwork(sc.AppID)) > 0 {
			summary["skipped"] = summary["skipped"].(int) + 1
			continue
		}
		var steamAppID int
		var name string
		m := matches.Get(sc.AppID)
		if m == nil {
			res, e := match.AutoMatch(sc.AppName, delistedApps())
			if e != nil {
				addErr(sc.AppName, e.Error())
				continue
			}
			if res == nil {
				summary["no_match"] = summary["no_match"].(int) + 1
				logger.Add(fmt.Sprintf("auto-match: sem correspondência para %s", sc.AppName))
				continue
			}
			steamAppID, name = res.AppID, res.Name
		} else {
			steamAppID, name = m.SteamAppID, m.Name
		}
		res, e := artwork.Download(sc.AppID, steamAppID, steamClient.Grid)
		if e != nil {
			addErr(sc.AppName, e.Error())
			continue
		}
		if len(res.Files) == 0 {
			addErr(sc.AppName, "nenhum asset baixado (CDN indisponível?)")
			continue
		}
		matches.Set(&store.Match{ShortcutAppID: sc.AppID, SteamAppID: steamAppID, Name: name})
		summary["applied"] = summary["applied"].(int) + 1
		logger.Add(fmt.Sprintf("auto-match: arte aplicada a %s (app %d)", name, steamAppID))
		time.Sleep(300 * time.Millisecond)
	}
	if err := matches.Save(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, summary)
}

// handleRemove apaga todos os arquivos de arte da grid de um atalho.
func handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST necessário", 405)
		return
	}
	var body struct {
		ShortcutAppID uint32 `json:"shortcut_appid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if _, err := findShortcut(body.ShortcutAppID); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	prefix := fmt.Sprintf("%d", body.ShortcutAppID)
	entries, err := os.ReadDir(steamClient.Grid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	removed := 0
	backupDir := filepath.Join(steamClient.Grid, "backup")
	_ = os.MkdirAll(backupDir, 0o755)
	for _, e := range entries {
		n := e.Name()
		if !strings.HasPrefix(n, prefix) {
			continue
		}
		rest := n[len(prefix):]
		if rest == "" || (rest[0] != 'p' && rest[0] != '_') {
			continue
		}
		src := filepath.Join(steamClient.Grid, n)
		dst := filepath.Join(backupDir, n)
		if err := os.Rename(src, dst); err == nil {
			removed++
		}
	}
	logger.Add(fmt.Sprintf("arte removida do atalho %d (%d arquivos)", body.ShortcutAppID, removed))
	writeJSON(w, map[string]any{"removed": removed})
}

func delistedApps() []delisted.App {
	if delistedIndex == nil {
		return nil
	}
	return delistedIndex.Apps
}

func findShortcut(appid uint32) (*steam.Shortcut, error) {
	list, err := steamClient.Shortcuts()
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].AppID == appid {
			return &list[i], nil
		}
	}
	return nil, fmt.Errorf("atalho %d não encontrado", appid)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func atoi(s string) int {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "darwin":
		cmd = "open"
	default:
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	}
	if cmd == "" {
		return
	}
	_ = exec.Command(cmd, append(args, url)...).Start()
}
