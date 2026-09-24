package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func get(t *testing.T, ctx context.Context, url string) (*http.Response, error) {
	t.Helper()
	return Do(&http.Client{Timeout: 5 * time.Second}, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		return req, nil
	})
}

func TestRepeteEm429ERecupera(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&chamadas, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	resp, err := get(t, context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&chamadas); got != 2 {
		t.Errorf("tentativas = %d, want 2", got)
	}
}

func TestErroPersistenteEsgotaTentativas(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&chamadas, 1)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	resp, err := get(t, context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Do deveria devolver a última resposta, não erro: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 (última resposta)", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&chamadas); got != tentativas {
		t.Errorf("tentativas = %d, want %d", got, tentativas)
	}
}

func TestNaoRepeteEmStatusFatal(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&chamadas, 1)
		http.Error(w, "não autorizado", http.StatusUnauthorized)
	}))
	defer srv.Close()

	resp, err := get(t, context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&chamadas); got != 1 {
		t.Errorf("tentativas = %d, want 1 (401 não é retentável)", got)
	}
}

func TestCancelaContextoDuranteBackoff(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&chamadas, 1)
		w.Header().Set("Retry-After", "10")
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	inicio := time.Now()
	_, err := get(t, ctx, srv.URL)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(inicio); elapsed > 2*time.Second {
		t.Errorf("cancelamento levou %v, queria sair do backoff na hora", elapsed)
	}
	if got := atomic.LoadInt32(&chamadas); got != 1 {
		t.Errorf("tentativas = %d, want 1", got)
	}
}

func TestErroDeRequestFactory(t *testing.T) {
	_, err := Do(http.DefaultClient, func() (*http.Request, error) {
		return nil, errors.New("url inválida")
	})
	if err == nil || !strings.Contains(err.Error(), "url inválida") {
		t.Fatalf("err = %v, want erro da factory", err)
	}
}

func TestErroDeTransporte(t *testing.T) {
	_, err := get(t, context.Background(), "http://127.0.0.1:1")
	if err == nil {
		t.Fatal("servidor inexistente deveria falhar")
	}
	if !strings.Contains(err.Error(), "tentativas") {
		t.Errorf("err = %v, quero menção a tentativas", err)
	}
}

func TestParseRetryAfter(t *testing.T) {
	casos := map[string]time.Duration{
		"":     0,
		"5":    5 * time.Second,
		"0":    0,
		"-1":   0,
		"data": 0,
		"http": 0,
		"120":  120 * time.Second,
	}
	for in, want := range casos {
		if got := parseRetryAfter(in); got != want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestDelayPrefereRetryAfterERespeitaTeto(t *testing.T) {
	if got := delay(1, 0); got != backoffBase {
		t.Errorf("delay(1) = %v, want %v", got, backoffBase)
	}
	if got := delay(1, 3*time.Second); got != 3*time.Second {
		t.Errorf("delay(1, 3s) = %v, want 3s", got)
	}
	if got := delay(3, 0); got > backoffMax {
		t.Errorf("delay(3) = %v, want <= %v", got, backoffMax)
	}
	if got := delay(1, time.Hour); got != retryMax {
		t.Errorf("delay(1, 1h) = %v, want teto %v", got, retryMax)
	}
}
