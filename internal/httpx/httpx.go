// Pacote httpx centraliza requisições HTTP com retries: 429/5xx e falhas de
// transporte são retentados com backoff exponencial, respeitando Retry-After
// e o contexto da requisição (cancelamento propaga na hora).
package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	tentativas  = 3
	backoffBase = 400 * time.Millisecond
	backoffMax  = 5 * time.Second
	retryMax    = 10 * time.Second
)

// retryable indica status transitórios que valem a pena tentar de novo.
func retryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusRequestTimeout,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

// Do executa mk() com até `tentativas` tentativas. Em sucesso (ou na última
// tentativa) devolve a resposta — o chamador fecha o body. mk é reexecutado
// a cada tentativa porque o body/conexão não podem ser reutilizados.
func Do(client *http.Client, mk func() (*http.Request, error)) (*http.Response, error) {
	var (
		lastErr    error
		retryAfter time.Duration
		ctx        = context.Background()
	)
	for attempt := 0; attempt < tentativas; attempt++ {
		if attempt > 0 {
			if err := sleep(ctx, delay(attempt, retryAfter)); err != nil {
				return nil, err
			}
		}
		req, err := mk()
		if err != nil {
			return nil, err
		}
		ctx = req.Context()
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if cerr := ctx.Err(); cerr != nil {
				return nil, cerr
			}
			continue
		}
		if !retryable(resp.StatusCode) || attempt == tentativas-1 {
			return resp, nil
		}
		retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 8<<10))
		resp.Body.Close()
	}
	if lastErr != nil {
		return nil, fmt.Errorf("falhou após %d tentativas: %w", tentativas, lastErr)
	}
	return nil, fmt.Errorf("falhou após %d tentativas", tentativas)
}

// delay calcula o backoff da tentativa `attempt` (1-based), preferindo o
// Retry-After quando a API informar.
func delay(attempt int, retryAfter time.Duration) time.Duration {
	d := backoffBase << (attempt - 1)
	if d > backoffMax {
		d = backoffMax
	}
	if retryAfter > d {
		d = retryAfter
	}
	if d > retryMax {
		d = retryMax
	}
	return d
}

// parseRetryAfter lê o header em segundos (a forma comum da SteamGridDB).
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(v)
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
