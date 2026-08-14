package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"steamart/internal/i18n"
	"strings"
)

func layoutSep() fyne.CanvasObject {
	s := widget.NewSeparator()
	s.Resize(fyne.NewSize(0, 12))
	return s
}

// card cria o card visual de um atalho na home.
func card(v shortcutView) fyne.CanvasObject {
	sc := v.Shortcut

	// título com ícone à esquerda
	title := widget.NewLabelWithStyle(sc.AppName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleRow := container.NewHBox()
	if icon := iconFor(sc); icon != nil {
		titleRow.Add(icon)
	}
	titleRow.Add(title)

	// caminho do executável
	exe := widget.NewLabelWithStyle(sc.Exe, fyne.TextAlignLeading, fyne.TextStyle{})

	// metadado do match
	meta := i18n.T("meta_none")
	if v.Match != nil {
		meta = i18n.T("meta", v.Match.Name, v.Match.SteamAppID)
	}
	metaLb := widget.NewLabelWithStyle(meta, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})

	// artes lado a lado com espaçamento uniforme
	artRow := container.NewHBox()
	for _, suf := range []string{"p", "", "_hero", "_logo", "_icon"} {
		if p := steamClient.GridPath(sc.AppID, suf); p != "" {
			img := canvas.NewImageFromFile(p)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(120, 68))
			artRow.Add(img)
		}
	}
	if len(artRow.Objects) == 0 {
		artRow.Add(widget.NewLabel(i18n.T("no_art")))
	}

	// botões de ação
	btns := container.NewHBox(
		widget.NewButton(i18n.T("auto_steam"), func() { doAutoSteam(v) }),
		widget.NewButton(i18n.T("details"), func() { doDetails(v) }),
		widget.NewButton(i18n.T("open_sgdb"), func() { doSGDB(v) }),
		widget.NewButton(i18n.T("remove_art"), func() { doClear(v) }),
	)

	return container.NewVBox(
		titleRow,
		exe,
		metaLb,
		artRow,
		btns,
		widget.NewSeparator(),
	)
}

// buildUI constrói a interface da janela principal.
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

	// barra de status + idioma + atualizar + sobre
	topRow := container.NewHBox(
		statusLb,
		layoutSpacer(),
		container.NewHBox(widget.NewLabel(i18n.T("language")), langSel),
		layoutSpacer(),
		refreshBtn,
		widget.NewButton(i18n.T("about"), func() { doAbout() }),
	)

	// chave SGDB
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

// renderList atualiza a lista de atalhos na home.
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
