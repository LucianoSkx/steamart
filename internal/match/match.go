package match

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"steamart/internal/delisted"
	"steamart/internal/title"
)

const (
	storeSearch = "https://store.steampowered.com/api/storesearch/?term=%s&l=brazilian&cc=BR"
	appDetails  = "https://store.steampowered.com/api/appdetails?appids=%d&l=brazilian&cc=BR"
)

var client = &http.Client{Timeout: 20 * time.Second}

// SearchResult é um resultado da busca na loja.
type SearchResult struct {
	AppID     int    `json:"id"`
	Name      string `json:"name"`
	TinyImage string `json:"tiny_image"`
}

// Meta é o conjunto de metadados ricos de um app Steam.
type Meta struct {
	AppID        int      `json:"appid"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Developers   []string `json:"developers"`
	Publishers   []string `json:"publishers"`
	ReleaseDate  string   `json:"release_date"`
	Genres       []string `json:"genres"`
	Categories   []string `json:"categories"`
	ShortDesc    string   `json:"short_desc"`
	HeaderImage  string   `json:"header_image"`
	Screenshots  []string `json:"screenshots"`
	Website      string   `json:"website"`
	DeckText     string   `json:"deck_text"`
	DeckCategory int      `json:"deck_category"`
}

func getJSON(url string, v any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "steamart/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d ao acessar %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// Search busca apps Steam por nome.
func Search(term string) ([]SearchResult, error) {
	url := fmt.Sprintf(storeSearch, url.QueryEscape(term))
	var out struct {
		Total int            `json:"total"`
		Items []SearchResult `json:"items"`
	}
	if err := getJSON(url, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// AutoMatch escolhe o melhor resultado de busca para um nome de atalho.
// Primeiro tenta a loja (descartando variantes não-primárias quando há
// alternativas); se nada casar bem, tenta o índice de jogos delisted.
func AutoMatch(term string, delistedApps []delisted.App) (*SearchResult, error) {
	results, err := Search(term)
	if err != nil {
		return nil, err
	}
	if best := pickBest(term, results); best != nil {
		return best, nil
	}
	if appid := delisted.ResolveAppID(term, delistedApps); appid != 0 {
		return &SearchResult{AppID: appid, Name: term}, nil
	}
	return nil, nil
}

// pickBest escolhe o melhor resultado entre os da loja: ignora variantes
// não-primárias (demo, beta, soundtrack, DLC...) quando existem alternativas
// e prioriza nomes com a mesma forma normalizada.
func pickBest(term string, results []SearchResult) *SearchResult {
	if len(results) == 0 {
		return nil
	}
	pool := results
	var primaries []SearchResult
	for _, r := range results {
		if !title.IsNonPrimaryTitle(r.Name) {
			primaries = append(primaries, r)
		}
	}
	if len(primaries) > 0 {
		pool = primaries
	}
	best := &pool[0]
	bestScore := matchScore(term, best.Name)
	for i := 1; i < len(pool); i++ {
		s := matchScore(term, pool[i].Name)
		if s > bestScore {
			bestScore = s
			best = &pool[i]
		}
	}
	if bestScore < 0.5 {
		return nil
	}
	return best
}

// matchScore mede a cobertura dos tokens normalizados da consulta no
// candidato (1 = mesma forma normalizada ou todos os tokens presentes).
// Em empates, a ordem da loja (relevância da Steam) decide.
func matchScore(term, name string) float64 {
	tn := title.NormaliseTitle(term)
	cn := title.NormaliseTitle(name)
	if tn == cn {
		return 1
	}
	wa := strings.Fields(tn)
	if len(wa) == 0 {
		return 0
	}
	set := map[string]bool{}
	for _, w := range strings.Fields(cn) {
		set[w] = true
	}
	hits := 0
	for _, w := range wa {
		if set[w] {
			hits++
		}
	}
	return float64(hits) / float64(len(wa))
}

// GetMeta busca os detalhes completos de um app.
func GetMeta(appid int) (*Meta, error) {
	url := fmt.Sprintf(appDetails, appid)
	var raw map[string]json.RawMessage
	if err := getJSON(url, &raw); err != nil {
		return nil, err
	}
	entry, ok := raw[fmt.Sprintf("%d", appid)]
	if !ok {
		return nil, fmt.Errorf("app %d sem dados", appid)
	}
	var obj struct {
		Success bool `json:"success"`
		Data    struct {
			Type       string   `json:"type"`
			Name       string   `json:"name"`
			Developers []string `json:"developers"`
			Publishers []string `json:"publishers"`
			Release    struct {
				Date string `json:"date"`
			} `json:"release_date"`
			Genres []struct {
				Description string `json:"description"`
			} `json:"genres"`
			Categories []struct {
				Description string `json:"description"`
			} `json:"categories"`
			ShortDesc   string `json:"short_description"`
			HeaderImage string `json:"header_image"`
			Screenshots []struct {
				Full string `json:"path_full"`
			} `json:"screenshots"`
			Website string `json:"website"`
			Deck    struct {
				Category    int    `json:"resolved_category"`
				Description string `json:"description"`
			} `json:"steam_deck_compatibility"`
		} `json:"data"`
	}
	if err := json.Unmarshal(entry, &obj); err != nil {
		return nil, err
	}
	if !obj.Success {
		return nil, fmt.Errorf("app %d sem sucesso na loja", appid)
	}
	m := &Meta{
		AppID:        appid,
		Name:         obj.Data.Name,
		Type:         obj.Data.Type,
		Developers:   obj.Data.Developers,
		Publishers:   obj.Data.Publishers,
		ReleaseDate:  obj.Data.Release.Date,
		ShortDesc:    obj.Data.ShortDesc,
		HeaderImage:  obj.Data.HeaderImage,
		Website:      obj.Data.Website,
		DeckText:     obj.Data.Deck.Description,
		DeckCategory: obj.Data.Deck.Category,
	}
	for _, g := range obj.Data.Genres {
		m.Genres = append(m.Genres, g.Description)
	}
	for _, c := range obj.Data.Categories {
		m.Categories = append(m.Categories, c.Description)
	}
	for _, s := range obj.Data.Screenshots {
		m.Screenshots = append(m.Screenshots, s.Full)
	}
	return m, nil
}
