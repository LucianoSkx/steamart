package sgdb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withServer(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	old := base
	base = srv.URL
	t.Cleanup(func() {
		base = old
		srv.Close()
	})
	return srv
}

func TestSearchEnviaAuthorization(t *testing.T) {
	var gotAuth, gotPath string
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    []map[string]any{{"id": 1, "name": "Portal"}},
		})
	})
	games, err := New("CHAVE-SECRETA").Search(context.Background(), "portal")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if gotAuth != "Bearer CHAVE-SECRETA" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if !strings.HasSuffix(gotPath, "/search/autocomplete/portal") {
		t.Errorf("path = %q", gotPath)
	}
	if len(games) != 1 || games[0].Name != "Portal" {
		t.Errorf("games = %+v", games)
	}
}

func TestImagesMontaFiltros(t *testing.T) {
	var gotQuery string
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    []map[string]any{{"id": 9, "url": "https://cdn/x.png"}},
		})
	})
	imgs, err := New("k").Images(context.Background(), 440, "hero", "920x430", true)
	if err != nil {
		t.Fatalf("Images: %v", err)
	}
	if !strings.Contains(gotQuery, "dimensions=920x430") || !strings.Contains(gotQuery, "mimes=") {
		t.Errorf("query = %q", gotQuery)
	}
	if len(imgs) != 1 {
		t.Errorf("imgs = %+v", imgs)
	}
}

func TestErroDeStatus(t *testing.T) {
	var chamadas int
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	})
	if _, err := New("k").Search(context.Background(), "x"); err == nil {
		t.Fatal("status 429 deveria virar erro")
	} else if !strings.Contains(err.Error(), "429") {
		t.Errorf("erro = %v, quero menção a 429", err)
	}
	if chamadas != 3 {
		t.Errorf("tentativas = %d, want 3 (retry em 429)", chamadas)
	}
}

func TestRetrySucessoNaSegundaTentativa(t *testing.T) {
	var chamadas int
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		if chamadas == 1 {
			http.Error(w, "temporário", http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    []map[string]any{{"id": 2, "name": "Portal 2"}},
		})
	})
	games, err := New("k").Search(context.Background(), "portal 2")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if chamadas != 2 {
		t.Errorf("tentativas = %d, want 2", chamadas)
	}
	if len(games) != 1 || games[0].Name != "Portal 2" {
		t.Errorf("games = %+v", games)
	}
}

func TestSucessoFalse(t *testing.T) {
	withServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false})
	})
	if _, err := New("k").Search(context.Background(), "x"); err == nil {
		t.Fatal("success=false deveria virar erro")
	}
}

func TestDownloadURLPrefereThumbNoIcone(t *testing.T) {
	var i Image
	i.URL = "https://cdn/a.ico"
	i.Thumb = "https://cdn/a.png"
	if got := i.DownloadURL("icon"); got != i.Thumb {
		t.Errorf("icon = %q, want thumb", got)
	}
	if got := i.DownloadURL("grid"); got != i.URL {
		t.Errorf("grid = %q, want url", got)
	}
}
