package main

import (
	"context"
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
	"steamart/internal/grid"
	"steamart/internal/httpx"
	"steamart/internal/i18n"
	"steamart/internal/icon"
	"steamart/internal/sgdb"
	"steamart/internal/steam"
	"steamart/internal/store"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

func fetchRemote(ctx context.Context, url string) ([]byte, error) {
	resp, err := httpx.Do(httpClient, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "steamart/1.0")
		return req, nil
	})
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
	steamClient *steam.Steam
	matches     *store.Store
	logger      *store.Logger

	// appCtx cancela downloads/buscas em voo quando o usuário sai do app.
	appCtx, appCancel = context.WithCancel(context.Background())

	// sgdbKey e delistedIndex são lidos por goroutines de busca e gravados
	// pela UI — protegidos por mutex embutidos.
	sgdbKey       sgdb.Key
	delistedIndex delisted.Holder

	appInstance fyne.App
	langSel     *widget.Select
	mainWin     fyne.Window
	statusLb    *widget.Label
	listBox     *fyne.Container
	listScr     *container.Scroll
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
	if matches.Recovered != "" {
		fmt.Fprintln(os.Stderr, "aviso:", matches.Recovered)
	}

	st := &store.Logger{}
	if err := st.SetFile(filepath.Join(c.Config, "steamart.log")); err == nil {
		logger = st
	}
	if matches.Recovered != "" {
		logger.Add("aviso: " + matches.Recovered)
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
	appCancel()
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
		sgdbKey.Set(k.Key)
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

// doClear remove toda a arte do atalho, movendo os arquivos para grid/backup/.
func doClear(v shortcutView) {
	dialog.ShowConfirm(i18n.T("remove_title"),
		i18n.T("remove_confirm"),
		func(ok bool) {
			if !ok {
				return
			}
			removed, err := grid.Remove(steamClient.Grid, v.Shortcut.AppID)
			if err != nil {
				dialog.ShowError(err, mainWin)
				return
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
