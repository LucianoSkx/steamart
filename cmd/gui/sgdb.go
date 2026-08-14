package main

import (
	"fmt"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"steamart/internal/artwork"
	"steamart/internal/i18n"
	"steamart/internal/official"
	"steamart/internal/sgdb"
)

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

	games := container.NewHBox()
	gamesScr := container.NewHScroll(games)
	gamesScr.SetMinSize(fyne.NewSize(720, 40))
	gamesScr.Hide()
	tabs := container.NewHBox()

	var currentGameID int
	currentAsset := "grid"
	currentSource := "sgdb"
	var temps []string
	var tempsMu sync.Mutex
	addTemp := func(p string) {
		tempsMu.Lock()
		defer tempsMu.Unlock()
		temps = append(temps, p)
	}
	applied := map[string]bool{}

	type galleryItem struct {
		im    sgdb.Image
		asset string
		a     *official.Asset
	}

	itemThumb := func(it galleryItem) string {
		if it.a != nil {
			return it.a.URL
		}
		return it.im.Thumb
	}
	itemApplyURL := func(it galleryItem) string {
		if it.a != nil {
			return it.a.URL
		}
		return it.im.DownloadURL(it.asset)
	}
	itemName := func(it galleryItem) string {
		if it.a != nil {
			return it.a.Name
		}
		return fmt.Sprintf("%s · %s", it.im.Author.Name, it.im.Style)
	}
	itemKind := func(it galleryItem) string {
		if it.a != nil {
			return it.a.Kind
		}
		return it.asset
	}

	imgBox := container.NewGridWrap(fyne.NewSize(240, 210))
	imgScroll := container.NewVScroll(imgBox)
	imgScroll.SetMinSize(fyne.NewSize(920, 520))

	makeCard := func(it galleryItem) fyne.CanvasObject {
		var img *canvas.Image
		if p, err := downloadTemp(itemThumb(it)); err == nil {
			addTemp(p)
			img = canvas.NewImageFromFile(p)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(220, 130))
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
				if _, err := artwork.SaveURL(itemApplyURL(it), steamClient.Grid, v.Shortcut.AppID, itemKind(it)); err != nil {
					ui(func() { dialog.ShowError(err, mainWin) })
					return
				}
				if logger != nil {
					if it.a != nil {
						logger.Add(i18n.T("official_log", itemKind(it), v.Shortcut.AppID))
					} else {
						logger.Add(i18n.T("sgdb_log", itemKind(it), v.Shortcut.AppID))
					}
				}
				ui(renderList)
				ui(func() {
					applied[itemThumb(it)] = true
					applyBtn.SetText(i18n.T("applied"))
					applyBtn.Disable()
				})
			}()
		})
		if applied[itemThumb(it)] {
			applyBtn.SetText(i18n.T("applied"))
			applyBtn.Disable()
		}
		return container.NewVBox(
			img,
			widget.NewLabel(itemName(it)),
			applyBtn,
		)
	}

	gen := 0
	loadInto := func(asset string) {
		currentAsset = asset
		gen++
		myGen := gen
		ui(func() { spinner.Show() })
		imgBox.Objects = imgBox.Objects[:0]
		imgBox.Refresh()
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
					})
					return
				}
				card := makeCard(galleryItem{a: picked})
				ui(func() {
					imgBox.Objects = imgBox.Objects[:0]
					imgBox.Add(card)
					imgBox.Refresh()
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
			})
			for _, im := range imgs {
				im := im
				go func() {
					card := makeCard(galleryItem{im: im, asset: asset})
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
				gamesScr.Show()
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
				gamesScr.Hide()
				anim.Hide()
				if v.Match != nil {
					gameLbl.SetText(i18n.T("official_for", v.Match.SteamAppID))
				}
			} else {
				if len(games.Objects) > 0 {
					gamesScr.Show()
				}
				anim.Show()
				gameLbl.SetText("")
			}
			loadInto(currentAsset)
		})
		srcBtns = append(srcBtns, btn)
		srcTabs.Add(btn)
	}
	srcBtns[0].Importance = widget.HighImportance

	// título + botão de fechar no topo
	var d dialog.Dialog
	titleRow := container.NewBorder(
		nil, nil, nil,
		widget.NewButton(i18n.T("close"), func() { d.Hide() }),
		widget.NewLabelWithStyle(i18n.T("open_sgdb"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	controls := container.NewVBox(
		header,
		container.NewBorder(nil, nil, nil, container.NewHBox(widget.NewButton(i18n.T("search"), run), anim), entry),
		srcTabs,
		gamesScr,
		container.NewHBox(gameLbl, layoutSpacer(), tabs),
		spinner,
	)
	box := container.NewBorder(container.NewVBox(titleRow, controls), nil, nil, nil, imgScroll)

	d = dialog.NewCustomWithoutButtons("", box, mainWin)
	d.Resize(fyne.NewSize(1024, 920))
	d.SetOnClosed(func() {
		tempsMu.Lock()
		defer tempsMu.Unlock()
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
