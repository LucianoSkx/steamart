package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"steamart/internal/delisted"
	"steamart/internal/i18n"
	"steamart/internal/icon"
	"steamart/internal/steam"
	"steamart/internal/store"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

func fetchRemote(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "steamart/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

const sgdbFile = "steamart-sgdb.json"

// Version é injetada no build via -ldflags="-X main.Version=v1.0.0"
var Version = "dev"

var (
	steamClient   *steam.Steam
	matches       *store.Store
	logger        *store.Logger
	sgdbKey       string
	delistedIndex *delisted.Index
	appInstance   fyne.App
	langSel       *widget.Select
	mainWin       fyne.Window
	statusLb      *widget.Label
	listBox       *fyne.Container
	listScr       *container.Scroll
)

type shortcutView struct {
	Shortcut steam.Shortcut
	Artwork  map[string]string
	Match    *store.Match
}

func ui(fn func()) {
	fyne.Do(fn)
}

func main() {
	c, err := steam.Discover()
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	steamClient = c

	matches, err = store.Open(c.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro ao abrir store:", err)
		os.Exit(1)
	}

	st := &store.Logger{}
	if err := st.SetFile(filepath.Join(c.Config, "steamart.log")); err == nil {
		logger = st
	}
	loadSGDBKey()

	a := app.NewWithID("com.steamart.app")
	if b, err := icon.PNG(512); err == nil {
		a.SetIcon(fyne.NewStaticResource("steamart.png", b))
	}
	appInstance = a
	prefLang := i18n.Lang(a.Preferences().StringWithFallback("lang", string(i18n.PTBR)))
	i18n.Set(prefLang)
	mainWin = a.NewWindow("SteamArt")

	setContent()
	mainWin.Resize(fyne.NewSize(900, 700))
	mainWin.ShowAndRun()
}

func setContent() {
	mainWin.SetContent(buildUI())
	renderList()
}

// layoutSpacer cria um espaço flexível horizontal ou vertical
func layoutSpacer() fyne.CanvasObject {
	return canvas.NewRectangle(&color.RGBA{0, 0, 0, 0})
}

func loadSGDBKey() {
	b, err := os.ReadFile(filepath.Join(steamClient.Config, sgdbFile))
	if err != nil {
		return
	}
	var k struct {
		Key string `json:"key"`
	}
	if json.Unmarshal(b, &k) == nil {
		sgdbKey = k.Key
	}
}

func saveSGDBKey(k string) {
	_ = steamClient.SaveJSON(sgdbFile, map[string]string{"key": k})
}

func loadShortcuts() []shortcutView {
	list, err := steamClient.Shortcuts()
	if err != nil {
		dialog.ShowError(err, mainWin)
		return nil
	}
	out := make([]shortcutView, 0, len(list))
	for _, sc := range list {
		out = append(out, shortcutView{
			Shortcut: sc,
			Artwork:  steamClient.HasArtwork(sc.AppID),
			Match:    matches.Get(sc.AppID),
		})
	}
	return out
}

func shortPath(p string) string {
	if h, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(p, h) {
			return "~" + strings.TrimPrefix(p, h)
		}
	}
	return p
}

func iconFor(sc steam.Shortcut) fyne.CanvasObject {
	// prefere o ícone da grid; cai no ícone do próprio atalho se for imagem.
	cands := []string{}
	if p := steamClient.GridPath(sc.AppID, "_icon"); p != "" {
		cands = append(cands, p)
	}
	if sc.Icon != "" {
		if ext := strings.ToLower(filepath.Ext(sc.Icon)); ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
			cands = append(cands, sc.Icon)
		}
	}
	for _, p := range cands {
		if _, err := os.Stat(p); err == nil {
			img := canvas.NewImageFromFile(p)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(32, 32))
			return img
		}
	}
	return nil
}

// belongsToShortcut indica se um arquivo da grid pertence ao atalho (pelo
// prefixo de appid), cobrindo capa (sem sufixo), grid (p) e demais sufixos (_).
func belongsToShortcut(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	rest := name[len(prefix):]
	if rest == "" {
		return false
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
	default:
		return false
	}
	return rest[0] == '.' || rest[0] == 'p' || rest[0] == '_'
}

// doClear remove toda a arte do atalho, movendo os arquivos para grid/backup/.
func doClear(v shortcutView) {
	dialog.ShowConfirm(i18n.T("remove_title"),
		i18n.T("remove_confirm"),
		func(ok bool) {
			if !ok {
				return
			}
			prefix := fmt.Sprintf("%d", v.Shortcut.AppID)
			entries, err := os.ReadDir(steamClient.Grid)
			if err != nil {
				dialog.ShowError(err, mainWin)
				return
			}
			bk := filepath.Join(steamClient.Grid, "backup")
			_ = os.MkdirAll(bk, 0o755)
			removed := 0
			for _, e := range entries {
				n := e.Name()
				if !belongsToShortcut(n, prefix) {
					continue
				}
				if err := os.Rename(filepath.Join(steamClient.Grid, n), filepath.Join(bk, n)); err == nil {
					removed++
				}
			}
			dialog.ShowInformation(i18n.T("remove_title"), i18n.T("removed", removed), mainWin)
			renderList()
		}, mainWin)
}

// doAbout abre o diálogo "Sobre" com versão, autor e links.
func doAbout() {
	info := fmt.Sprintf("%s v%s\n\n%s\n\n%s",
		i18n.T("app_name"),
		Version,
		i18n.T("about_desc"),
		i18n.T("about_links"),
	)
	dialog.ShowInformation(i18n.T("about_title"), info, mainWin)
}
