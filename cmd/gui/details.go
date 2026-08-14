package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"steamart/internal/i18n"
	"steamart/internal/match"
)

const metaCacheFile = "steamart-meta.json"

type cachedMeta struct {
	FetchedAt time.Time  `json:"fetched_at"`
	Meta      match.Meta `json:"meta"`
}

// doDetails abre a janela de detalhes (Game Info) do jogo casado, usando
// metadados da loja com cache local de 7 dias.
func doDetails(v shortcutView) {
	if v.Match == nil {
		dialog.ShowInformation(i18n.T("warning"), i18n.T("need_match_details"), mainWin)
		return
	}
	cleanupStaleTemps()
	spinner := widget.NewProgressBarInfinite()
	spinner.Start()
	var d dialog.Dialog
	titleRow := container.NewBorder(
		nil, nil, nil,
		widget.NewButton(i18n.T("close"), func() { d.Hide() }),
		widget.NewLabelWithStyle(i18n.T("details_title", v.Shortcut.AppName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
	box := container.NewVBox(spinner)
	d = dialog.NewCustomWithoutButtons("", box, mainWin)
	d.Resize(fyne.NewSize(760, 720))
	var temps []string
	var tempsMu sync.Mutex
	d.SetOnClosed(func() {
		tempsMu.Lock()
		defer tempsMu.Unlock()
		for _, p := range temps {
			_ = os.Remove(p)
		}
	})
	d.Show()

	go func() {
		meta := loadCachedMeta(v.Match.SteamAppID)
		if meta == nil {
			m, err := match.GetMeta(v.Match.SteamAppID)
			if err != nil {
				ui(func() {
					spinner.Stop()
					box.Objects = []fyne.CanvasObject{
						titleRow,
						widget.NewLabel(i18n.T("no_store_page")),
					}
					box.Refresh()
				})
				return
			}
			meta = m
			saveCachedMeta(m)
		}
		content, newTemps := detailsContent(meta)
		tempsMu.Lock()
		temps = append(temps, newTemps...)
		tempsMu.Unlock()
		ui(func() {
			spinner.Stop()
			box.Objects = []fyne.CanvasObject{titleRow, content}
			box.Refresh()
		})
	}()
}

func detailsContent(meta *match.Meta) (fyne.CanvasObject, []string) {
	var temps []string
	rows := container.NewVBox()

	if meta.HeaderImage != "" {
		if p, err := downloadTemp(meta.HeaderImage); err == nil {
			temps = append(temps, p)
			img := canvas.NewImageFromFile(p)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(460, 215))
			rows.Add(img)
		}
	}
	if meta.ShortDesc != "" {
		lb := widget.NewLabel(meta.ShortDesc)
		lb.Wrapping = fyne.TextWrapWord
		rows.Add(lb)
	}

	add := func(label, value string) {
		if value == "" {
			return
		}
		rows.Add(widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		lb := widget.NewLabel(value)
		lb.Wrapping = fyne.TextWrapWord
		rows.Add(lb)
	}
	add(i18n.T("meta_devs"), strings.Join(meta.Developers, ", "))
	add(i18n.T("meta_pubs"), strings.Join(meta.Publishers, ", "))
	add(i18n.T("meta_release"), meta.ReleaseDate)
	add(i18n.T("meta_genres"), strings.Join(meta.Genres, ", "))
	add(i18n.T("meta_cats"), strings.Join(meta.Categories, ", "))
	if meta.DeckText != "" {
		add(i18n.T("meta_deck"), meta.DeckText)
	}
	if meta.Website != "" {
		rows.Add(container.NewHBox(
			widget.NewLabel(i18n.T("meta_website")),
			widget.NewHyperlink(meta.Website, mustURL(meta.Website)),
		))
	}

	if len(meta.Screenshots) > 0 {
		rows.Add(widget.NewLabelWithStyle(i18n.T("meta_screens"), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		shots := container.NewGridWrap(fyne.NewSize(330, 186))
		for i, s := range meta.Screenshots {
			if i >= 6 {
				break
			}
			if p, err := downloadTemp(s); err == nil {
				temps = append(temps, p)
				img := canvas.NewImageFromFile(p)
				img.FillMode = canvas.ImageFillContain
				img.SetMinSize(fyne.NewSize(330, 186))
				shots.Add(img)
			}
		}
		rows.Add(shots)
	}

	scr := container.NewVScroll(rows)
	scr.SetMinSize(fyne.NewSize(720, 640))
	return scr, temps
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		return nil
	}
	return u
}

// loadCachedMeta devolve os metadados em cache para um app, se ainda
// estiverem dentro do TTL de 7 dias.
func loadCachedMeta(appid int) *match.Meta {
	var m map[string]cachedMeta
	if steamClient.LoadJSON(metaCacheFile, &m) != nil {
		return nil
	}
	c, ok := m[fmt.Sprintf("%d", appid)]
	if !ok || time.Since(c.FetchedAt) > 7*24*time.Hour {
		return nil
	}
	return &c.Meta
}

// saveCachedMeta grava os metadados de um app no cache local.
func saveCachedMeta(meta *match.Meta) {
	var m map[string]cachedMeta
	_ = steamClient.LoadJSON(metaCacheFile, &m)
	if m == nil {
		m = map[string]cachedMeta{}
	}
	m[fmt.Sprintf("%d", meta.AppID)] = cachedMeta{FetchedAt: time.Now(), Meta: *meta}
	_ = steamClient.SaveJSON(metaCacheFile, m)
}
