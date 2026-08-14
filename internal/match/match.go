package match

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
func AutoMatch(term string) (*SearchResult, error) {
	results, err := Search(term)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	best := results[0]
	bestScore := score(term, best.Name)
	for _, r := range results[1:] {
		s := score(term, r.Name)
		if s > bestScore {
			bestScore = s
			best = r
		}
	}
	if bestScore < 0.5 {
		return nil, nil
	}
	return &best, nil
}

func score(a, b string) float64 {
	na := norm(a)
	nb := norm(b)
	if na == nb {
		return 1
	}
	wa := strings.Fields(na)
	wb := strings.Fields(nb)
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	set := map[string]bool{}
	for _, w := range wa {
		set[w] = true
	}
	hits := 0
	for _, w := range wb {
		if set[w] {
			hits++
		}
	}
	return float64(hits) / float64(len(wb))
}

func norm(s string) string {
	r := strings.NewReplacer(":", "", "-", "", ".", "", "'", "", "!", "", "&", "and")
	s = r.Replace(strings.ToLower(s))
	return strings.Join(strings.Fields(s), " ")
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
