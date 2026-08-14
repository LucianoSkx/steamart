// Pacote delisted baixa e consulta o índice de jogos removidos da loja
// Steam (steam-tracker.com), para casar títulos que a busca da loja não
// encontra mais. Portado do Decky-Metadata (backend/providers/delisted.py,
// GPL-3.0).
package delisted

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"steamart/internal/title"
)

const (
	// SourceURL é a página com a lista de jogos delisted.
	SourceURL = "https://steam-tracker.com/apps/delisted"
	// TTL define a validade do índice em cache (7 dias).
	TTL = 7 * 24 * time.Hour
	// maxBytes limita o tamanho do HTML baixado.
	maxBytes = 30 << 20
	// minApps é o mínimo de apps esperado para aceitar o download.
	minApps = 100
)

var appLinkRe = regexp.MustCompile(`(?i)href='https://steam-tracker\.com/app/(\d+)/'[^>]*>\s*([^<]+?)\s*</a>`)

// App é um jogo delisted com seu appid original na Steam.
type App struct {
	AppID int    `json:"appid"`
	Name  string `json:"name"`
}

// Index é o catálogo de jogos delisted.
type Index struct {
	FetchedAt time.Time `json:"fetched_at"`
	Source    string    `json:"source"`
	Apps      []App     `json:"apps"`
}

// Fresh indica se o índice ainda está dentro do TTL.
func (i *Index) Fresh() bool {
	return i != nil && !i.FetchedAt.IsZero() && time.Since(i.FetchedAt) < TTL
}

// Ensure devolve um índice fresco: o da memória, o do disco ou um recém
// baixado (gravado em cache). Se tudo falhar, devolve o melhor disponível
// (pode ser nil).
func Ensure(mem *Index, cachePath string, force bool) *Index {
	if mem != nil && !force && mem.Fresh() {
		return mem
	}
	if idx := Load(cachePath); idx != nil && !force && idx.Fresh() {
		return idx
	}
	if idx, err := Download(); err == nil {
		_ = Save(cachePath, idx)
		return idx
	}
	if idx := Load(cachePath); idx != nil {
		return idx
	}
	return mem
}

// Download baixa e parseia o índice de steam-tracker.com.
func Download() (*Index, error) {
	req, err := http.NewRequest(http.MethodGet, SourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "steamart/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d ao baixar %s", resp.StatusCode, SourceURL)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBytes {
		return nil, fmt.Errorf("índice delisted excede %d bytes", maxBytes)
	}
	apps := parseHTML(body)
	if len(apps) < minApps {
		return nil, fmt.Errorf("índice delisted implausível: %d apps", len(apps))
	}
	return &Index{FetchedAt: time.Now(), Source: SourceURL, Apps: apps}, nil
}

func parseHTML(text []byte) []App {
	var out []App
	seen := map[int]bool{}
	for _, m := range appLinkRe.FindAllSubmatch(text, -1) {
		appid, err := strconv.Atoi(string(m[1]))
		name := title.CleanTitle(html.UnescapeString(string(m[2])))
		if err != nil || appid == 0 || name == "" || seen[appid] {
			continue
		}
		seen[appid] = true
		out = append(out, App{AppID: appid, Name: name})
	}
	return out
}

// Load lê o índice do cache em disco.
func Load(path string) *Index {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var idx Index
	if json.Unmarshal(b, &idx) != nil {
		return nil
	}
	if len(idx.Apps) == 0 {
		return nil
	}
	return &idx
}

// Save grava o índice no cache em disco (atômico via arquivo temporário).
func Save(path string, idx *Index) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ResolveAppID devolve o appid do jogo delisted que melhor casa com o
// título (0 se nada casar bem o bastante).
func ResolveAppID(name string, apps []App) int {
	clean := title.CleanTitle(name)
	if clean == "" {
		return 0
	}
	query := title.NormaliseTitle(clean)
	if query == "" {
		return 0
	}
	queryNums := numberSet(query)
	type cand struct {
		score int
		appid int
		name  string
	}
	var best *cand
	for _, a := range apps {
		if a.AppID == 0 || a.Name == "" {
			continue
		}
		candidate := title.NormaliseTitle(a.Name)
		if candidate == "" || !title.DistinctiveTokensPresent(query, candidate) {
			continue
		}
		score := 0
		if candidate == query {
			score = 1000
		} else {
			ratio := title.LevRatio(query, candidate)
			if ratio < 0.72 {
				continue
			}
			score = int(ratio * 500)
		}
		if title.IsNonPrimaryTitle(a.Name) {
			score -= 800
		}
		if diff := numberSet(candidate); !subset(diff, queryNums) {
			score -= 120
		}
		if best == nil || score > best.score {
			best = &cand{score, a.AppID, a.Name}
		}
	}
	if best == nil || best.score < 300 {
		return 0
	}
	return best.appid
}

func numberSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, m := range regexp.MustCompile(`\d+`).FindAllString(s, -1) {
		out[m] = true
	}
	return out
}

func subset(a, b map[string]bool) bool {
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
