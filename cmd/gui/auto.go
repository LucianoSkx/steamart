package main

import (
	"path/filepath"

	"fyne.io/fyne/v2/dialog"
	"steamart/internal/artwork"
	"steamart/internal/delisted"
	"steamart/internal/i18n"
	"steamart/internal/match"
	"steamart/internal/store"
)

// doAutoSteam casa o atalho com a Steam e aplica as artes oficiais da loja
// (grid, hero, logo, capsule) automaticamente, sem escolha manual.
// Se a loja não achar, tenta o índice de jogos delisted.
func doAutoSteam(v shortcutView) {
	go func() {
		if idx := delisted.Ensure(delistedIndex, filepath.Join(steamClient.Config, "delisted_index.json"), false); idx != nil {
			delistedIndex = idx
		}
		var apps []delisted.App
		if delistedIndex != nil {
			apps = delistedIndex.Apps
		}
		r, err := match.AutoMatch(v.Shortcut.AppName, apps)
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
