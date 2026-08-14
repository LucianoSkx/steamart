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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"steamart/internal/artwork"
	"steamart/internal/i18n"
	"steamart/internal/icon"
	"steamart/internal/match"
	"steamart/internal/official"
	"steamart/internal/sgdb"
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

var (
	steamClient *steam.Steam
	matches     *store.Store
	logger      *store.Logger
	sgdbKey     string
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

func buildUI() fyne.CanvasObject {
	statusLb = widget.NewLabel("")

	sgdbEntry := widget.NewPasswordEntry()
	sgdbEntry.SetPlaceHolder(i18n.T("key_placeholder"))
	if sgdbKey != "" {
		sgdbEntry.SetText(sgdbKey)
	}
	saveKeyBtn := widget.NewButton(i18n.T("save_key"), func() {
		k := strings.TrimSpace(sgdbEntry.Text)
		if k == "" || k == "********" {
			dialog.ShowInformation(i18n.T("warning"), i18n.T("key_empty"), mainWin)
			return
		}
		saveSGDBKey(k)
		sgdbKey = k
		sgdbEntry.SetText(k)
		dialog.ShowInformation(i18n.T("ok"), i18n.T("key_saved"), mainWin)
	})

	refreshBtn := widget.NewButton(i18n.T("refresh"), func() { renderList() })

	langSel = widget.NewSelect([]string{i18n.LangName(i18n.PTBR), i18n.LangName(i18n.EN)}, nil)
	langSel.SetSelected(i18n.LangName(i18n.Get()))
	langSel.OnChanged = func(name string) {
		var l i18n.Lang
		for _, cand := range i18n.Languages() {
			if i18n.LangName(cand) == name {
				l = cand
			}
		}
		i18n.Set(l)
		appInstance.Preferences().SetString("lang", string(l))
		setContent()
	}

	topRow := container.NewHBox(
		statusLb,
		layoutSpacer(),
		container.NewHBox(widget.NewLabel(i18n.T("language")), langSel),
		refreshBtn,
	)
	keyRow := container.NewBorder(
		nil,
		nil,
		widget.NewLabel(i18n.T("key_label")),
		saveKeyBtn,
		sgdbEntry,
	)

	top := container.NewVBox(topRow, keyRow)

	listBox = container.NewVBox()
	listScr = container.NewVScroll(listBox)
	return container.NewBorder(top, nil, nil, nil, listScr)
}

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

func renderList() {
	views := loadShortcuts()
	listBox.Objects = listBox.Objects[:0]
	pending := 0
	for _, v := range views {
		if v.Match == nil || len(v.Artwork) == 0 {
			pending++
		}
		listBox.Add(card(v))
	}
	if len(views) == 0 {
		listBox.Add(widget.NewLabel(i18n.T("no_shortcuts")))
	}
	statusLb.SetText(i18n.T("status", shortPath(steamClient.Root), len(views), pending))
	listBox.Refresh()
	listScr.Refresh()
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

func card(v shortcutView) fyne.CanvasObject {
	sc := v.Shortcut
	title := widget.NewLabelWithStyle(sc.AppName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleRow := container.NewHBox()
	if icon := iconFor(sc); icon != nil {
		titleRow.Add(icon)
	}
	titleRow.Add(title)
	exe := widget.NewLabel(sc.Exe)

	meta := i18n.T("meta_none")
	if v.Match != nil {
		meta = i18n.T("meta", v.Match.Name, v.Match.SteamAppID)
	}
	metaLb := widget.NewLabel(meta)

	// miniaturas da arte existente
	thumbs := container.NewHBox()
	for _, suf := range []string{"p", "_hero", "_logo", "_icon", ""} {
		if p := steamClient.GridPath(sc.AppID, suf); p != "" {
			img := canvas.NewImageFromFile(p)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(70, 100))
			thumbs.Add(img)
		}
	}
	if len(thumbs.Objects) == 0 {
		thumbs.Add(widget.NewLabel(i18n.T("no_art")))
	}

	btns := container.NewHBox(
		widget.NewButton(i18n.T("auto_steam"), func() { doAutoSteam(v) }),
		widget.NewButton(i18n.T("open_sgdb"), func() { doSGDB(v) }),
		widget.NewButton(i18n.T("remove_art"), func() { doClear(v) }),
	)

	return container.NewVBox(
		titleRow,
		exe,
		metaLb,
		thumbs,
		btns,
		widget.NewSeparator(),
	)
}

// doAutoSteam casa o atalho com a Steam e aplica as artes oficiais da loja
// (grid, hero, logo, capsule) automaticamente, sem escolha manual.
func doAutoSteam(v shortcutView) {
	go func() {
		r, err := match.AutoMatch(v.Shortcut.AppName)
		if err != nil {
			ui(func() { dialog.ShowError(err, mainWin) })
			return
		}
		if r == nil {
			ui(func() {
				dialog.ShowInformation(i18n.T("auto_steam"), i18n.T("auto_no_match", v.Shortcut.AppName), mainWin)
			})
			return
		}
		applySteam(v, r.AppID, r.Name)
	}()
}

// applySteam baixa as artes oficiais da loja Steam e as grava na grid do atalho.
func applySteam(v shortcutView, steamAppID int, name string) {
	go func() {
		res, err := artwork.Download(v.Shortcut.AppID, steamAppID, steamClient.Grid)
		if err != nil {
			ui(func() { dialog.ShowError(err, mainWin) })
			return
		}
		if len(res.Files) == 0 {
			ui(func() { dialog.ShowInformation(i18n.T("warning"), i18n.T("cdn_unavailable"), mainWin) })
			return
		}
		matches.Set(&store.Match{ShortcutAppID: v.Shortcut.AppID, SteamAppID: steamAppID, Name: name})
		if err := matches.Save(); err != nil {
			ui(func() { dialog.ShowError(err, mainWin) })
			return
		}
		if logger != nil {
			logger.Add(i18n.T("log_art", name, steamAppID))
		}
		ui(renderList)
	}()
}


func doSGDB(v shortcutView) {
	if sgdbKey == "" {
		dialog.ShowInformation(i18n.T("warning"), i18n.T("sgdb_key_missing"), mainWin)
		return
	}
	cleanupStaleTemps()
	entry := widget.NewEntry()
	entry.SetPlaceHolder(i18n.T("sgdb_search_ph"))
	anim := widget.NewCheck(i18n.T("animated_only"), nil)

	header := widget.NewLabelWithStyle(i18n.T("sgdb_for", v.Shortcut.AppName), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	gameLbl := widget.NewLabel("")
	spinner := widget.NewProgressBarInfinite()
	spinner.Hide()

	games := container.NewVBox()
	tabs := container.NewHBox()
	imgBox := container.NewGridWrap(fyne.NewSize(210, 190))
	imgScroll := container.NewVScroll(imgBox)
	imgScroll.SetMinSize(fyne.NewSize(720, 430))

	var currentGameID int
	currentAsset := "grid"
	currentSource := "sgdb"
	var temps []string

	imageCard := func(im sgdb.Image, asset string) fyne.CanvasObject {
		var img *canvas.Image
		if p, err := downloadTemp(im.Thumb); err == nil {
			temps = append(temps, p)
			img = canvas.NewImageFromFile(p)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(190, 110))
		} else {
			img = canvas.NewImageFromResource(theme.BrokenImageIcon())
		}
		var applyBtn *widget.Button
		applyBtn = widget.NewButton(i18n.T("apply"), func() {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						ui(func() { dialog.ShowError(fmt.Errorf("erro ao aplicar: %v", r), mainWin) })
					}
				}()
				if _, err := artwork.SaveURL(im.DownloadURL(asset), steamClient.Grid, v.Shortcut.AppID, asset); err != nil {
					ui(func() { dialog.ShowError(err, mainWin) })
					return
				}
				if logger != nil {
					logger.Add(i18n.T("sgdb_log", asset, v.Shortcut.AppID))
				}
				ui(renderList)
				ui(func() {
					applyBtn.SetText(i18n.T("applied"))
					applyBtn.Disable()
				})
			}()
		})
		return container.NewVBox(
			img,
			widget.NewLabel(fmt.Sprintf("%s · %s", im.Author.Name, im.Style)),
			applyBtn,
		)
	}

	officialCard := func(a official.Asset) fyne.CanvasObject {
		var img *canvas.Image
		if p, err := downloadTemp(a.URL); err == nil {
			temps = append(temps, p)
			img = canvas.NewImageFromFile(p)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(190, 110))
		} else {
			img = canvas.NewImageFromResource(theme.BrokenImageIcon())
		}
		var applyBtn *widget.Button
		applyBtn = widget.NewButton(i18n.T("apply"), func() {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						ui(func() { dialog.ShowError(fmt.Errorf("erro ao aplicar: %v", r), mainWin) })
					}
				}()
				if _, err := artwork.SaveURL(a.URL, steamClient.Grid, v.Shortcut.AppID, a.Kind); err != nil {
					ui(func() { dialog.ShowError(err, mainWin) })
					return
				}
				if logger != nil {
					logger.Add(i18n.T("official_log", a.Kind, v.Shortcut.AppID))
				}
				ui(renderList)
				ui(func() {
					applyBtn.SetText(i18n.T("applied"))
					applyBtn.Disable()
				})
			}()
		})
		return container.NewVBox(img, widget.NewLabel(a.Name), applyBtn)
	}

	gen := 0
	loadInto := func(asset string) {
		currentAsset = asset
		gen++
		myGen := gen
		ui(func() { spinner.Show() })
		if currentSource == "official" {
			go func() {
				var picked *official.Asset
				if v.Match != nil {
					assets := official.List(v.Match.SteamAppID)
					for i := range assets {
						if assets[i].Kind == asset {
							picked = &assets[i]
							break
						}
					}
				}
				ui(func() {
					if myGen == gen {
						spinner.Hide()
					}
				})
				if myGen != gen {
					return
				}
				if picked == nil {
					ui(func() {
						imgBox.Objects = imgBox.Objects[:0]
						imgBox.Add(widget.NewLabel(i18n.T("official_unavail")))
						imgBox.Refresh()
						imgScroll.Refresh()
					})
					return
				}
				card := officialCard(*picked)
				ui(func() {
					imgBox.Objects = imgBox.Objects[:0]
					imgBox.Add(card)
					imgBox.Refresh()
					imgScroll.Refresh()
				})
			}()
			return
		}
		gameID := currentGameID
		go func() {
			apiAsset, dims := asset, ""
			if asset == "capsule" {
				apiAsset, dims = "grid", "920x430"
			}
			imgs, err := sgdb.New(sgdbKey).Images(gameID, apiAsset, dims, anim.Checked)
			ui(func() {
				if myGen == gen {
					spinner.Hide()
				}
			})
			if myGen != gen {
				return
			}
			if err != nil {
				ui(func() { dialog.ShowError(err, mainWin) })
				return
			}
			ui(func() {
				imgBox.Objects = imgBox.Objects[:0]
				if len(imgs) == 0 {
					imgBox.Add(widget.NewLabel(i18n.T("no_images")))
				}
				imgBox.Refresh()
				imgScroll.Refresh()
			})
			for _, im := range imgs {
				im := im
				go func() {
					card := imageCard(im, asset)
					ui(func() {
						if myGen != gen {
							return
						}
						imgBox.Add(card)
						imgBox.Refresh()
					})
				}()
			}
		}()
	}

	showImages := func(gameID int, name string) {
		currentGameID = gameID
		gameLbl.SetText(i18n.T("sgdb_showing", name))
		tabs.Objects = tabs.Objects[:0]
		for _, a := range []struct {
			label, key string
		}{
			{i18n.T("tab_grid"), "grid"},
			{i18n.T("tab_hero"), "hero"},
			{i18n.T("tab_logo"), "logo"},
			{i18n.T("tab_icon"), "icon"},
			{i18n.T("tab_capsule"), "capsule"},
		} {
			a := a
			var btn *widget.Button
			btn = widget.NewButton(a.label, func() {
				for _, b := range tabs.Objects {
					if x, ok := b.(*widget.Button); ok {
						x.Importance = widget.MediumImportance
					}
				}
				btn.Importance = widget.HighImportance
				loadInto(a.key)
			})
			tabs.Add(btn)
		}
		tabs.Refresh()
		loadInto("grid")
	}

	anim.OnChanged = func(bool) { loadInto(currentAsset) }

	run := func() {
		go func() {
			gs, err := sgdb.New(sgdbKey).Search(entry.Text)
			if err != nil {
				ui(func() { dialog.ShowError(err, mainWin) })
				return
			}
			ui(func() {
				games.Objects = games.Objects[:0]
				if len(gs) == 0 {
					games.Add(widget.NewLabel(i18n.T("nothing_found")))
				}
				for _, g := range gs {
					g := g
					games.Add(widget.NewButton(g.Name, func() {
						showImages(g.ID, g.Name)
					}))
				}
				games.Refresh()
				if len(gs) > 0 {
					showImages(gs[0].ID, gs[0].Name)
				}
			})
		}()
	}
	entry.OnSubmitted = func(string) { run() }

	srcTabs := container.NewHBox()
	var srcBtns []*widget.Button
	for _, s := range []struct {
		label, key string
	}{
		{i18n.T("src_sgdb"), "sgdb"},
		{i18n.T("src_official"), "official"},
	} {
		s := s
		var btn *widget.Button
		btn = widget.NewButton(s.label, func() {
			if s.key == "official" && v.Match == nil {
				dialog.ShowInformation(i18n.T("warning"), i18n.T("need_match"), mainWin)
				return
			}
			currentSource = s.key
			for _, b := range srcBtns {
				b.Importance = widget.MediumImportance
			}
			btn.Importance = widget.HighImportance
			if s.key == "official" {
				games.Hide()
				anim.Hide()
				if v.Match != nil {
					gameLbl.SetText(i18n.T("official_for", v.Match.SteamAppID))
				}
			} else {
				games.Show()
				anim.Show()
				gameLbl.SetText("")
			}
			loadInto(currentAsset)
		})
		srcBtns = append(srcBtns, btn)
		srcTabs.Add(btn)
	}
	srcBtns[0].Importance = widget.HighImportance

	box := container.NewVBox(
		header,
		entry,
		container.NewHBox(layoutSpacer(), widget.NewButton(i18n.T("search"), run), anim),
		widget.NewLabel(i18n.T("search_hint")),
		srcTabs,
		gameLbl,
		spinner,
		games,
		tabs,
		imgScroll,
	)

	// diálogo
	d := dialog.NewCustom(i18n.T("open_sgdb"), i18n.T("close"), box, mainWin)
	d.Resize(fyne.NewSize(760, 720))
	d.SetOnClosed(func() {
		for _, p := range temps {
			_ = os.Remove(p)
		}
	})
	d.Show()

	// pré-carrega o jogo do atalho (sem busca manual)
	if v.Match != nil {
		go func() {
			g, err := sgdb.New(sgdbKey).GameBySteamAppID(v.Match.SteamAppID)
			if err != nil {
				ui(func() { entry.SetText(v.Shortcut.AppName); run() })
				return
			}
			ui(func() { showImages(g.ID, g.Name) })
		}()
	} else {
		entry.SetText(v.Shortcut.AppName)
		run()
	}
}

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


