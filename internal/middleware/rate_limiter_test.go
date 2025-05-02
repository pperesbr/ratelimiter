package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pperesbr/ratelimiter/config"
	"github.com/pperesbr/ratelimiter/internal/ratelimiter"
	"github.com/pperesbr/ratelimiter/internal/ratelimiter/store"
)

func TestRateLimiterMiddleware(t *testing.T) {
	// Cria configuração de teste
	cfg := &config.Config{
		RateLimitIP:             3,
		RateLimitIPBlockTime:    5 * time.Minute,
		RateLimitToken:          5,
		RateLimitTokenBlockTime: 5 * time.Minute,
	}

	// Factory para armazenamento em memória
	storeFactory := func() (store.RateLimiterStore, error) {
		return store.NewMemoryStore(), nil
	}

	// Cria rate limiter
	limiter, err := ratelimiter.NewRateLimiter(cfg, storeFactory)
	if err != nil {
		t.Fatalf("falha ao criar rate limiter: %v", err)
	}
	defer limiter.Close()

	// Cria middleware
	middleware := NewRateLimiterMiddleware(limiter)

	// Handler de teste simples
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Handler final com middleware
	handler := middleware.Middleware(nextHandler)

	// ---- Teste 1: Limite por IP ----
	t.Run("IP Limit", func(t *testing.T) {
		// Cria um servidor de teste
		server := httptest.NewServer(handler)
		defer server.Close()

		// Envia requisições até o limite
		for i := 0; i < cfg.RateLimitIP; i++ {
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("erro ao fazer requisição: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("esperava status %d, mas recebeu %d", http.StatusOK, resp.StatusCode)
			}
			resp.Body.Close()
		}

		// A próxima requisição deve ser bloqueada
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("erro ao fazer requisição: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("esperava status %d, mas recebeu %d", http.StatusTooManyRequests, resp.StatusCode)
		}
	})

	// ---- Teste 2: Limite por Token ----
	t.Run("Token Limit", func(t *testing.T) {
		// Cria um novo rate limiter para este teste
		limiter, err := ratelimiter.NewRateLimiter(cfg, storeFactory)
		if err != nil {
			t.Fatalf("falha ao criar rate limiter: %v", err)
		}
		defer limiter.Close()

		// Cria novo middleware
		middleware := NewRateLimiterMiddleware(limiter)
		handler := middleware.Middleware(nextHandler)

		// Cria um servidor de teste
		server := httptest.NewServer(handler)
		defer server.Close()

		// Cria um cliente HTTP para manter cookies/headers
		client := &http.Client{}

		// Envia requisições com token até o limite
		for i := 0; i < cfg.RateLimitToken; i++ {
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatalf("erro ao criar requisição: %v", err)
			}

			// Adiciona o token no header
			req.Header.Add("API_KEY", "test-token")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("erro ao fazer requisição: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("esperava status %d, mas recebeu %d", http.StatusOK, resp.StatusCode)
			}
			resp.Body.Close()
		}

		// A próxima requisição deve ser bloqueada
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatalf("erro ao criar requisição: %v", err)
		}
		req.Header.Add("API_KEY", "test-token")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("erro ao fazer requisição: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("esperava status %d, mas recebeu %d", http.StatusTooManyRequests, resp.StatusCode)
		}
	})
}
