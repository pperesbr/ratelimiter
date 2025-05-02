package middleware

import (
	"net"
	"net/http"

	"github.com/pperesbr/ratelimiter/internal/ratelimiter"
)

// RateLimiterMiddleware é um middleware para limitar requisições
type RateLimiterMiddleware struct {
	limiter *ratelimiter.RateLimiter
}

// NewRateLimiterMiddleware cria uma nova instância do middleware
func NewRateLimiterMiddleware(limiter *ratelimiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: limiter,
	}
}

// Middleware retorna uma função de middleware HTTP
func (m *RateLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrai o IP real do cliente (considerando cabeçalhos de proxy)
		ip := getClientIP(r)

		// Extrai o token de acesso do cabeçalho (se presente)
		token := r.Header.Get("API_KEY")

		// Cria uma requisição para o rate limiter
		req := &ratelimiter.LimiterRequest{
			IP:    ip,
			Token: token,
		}

		// Verifica se a requisição deve ser permitida
		if err := m.limiter.Allow(r.Context(), req); err != nil {
			// Se o limite foi excedido, retorna 429 Too Many Requests
			if _, ok := err.(*ratelimiter.LimitExceededError); ok {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}

			// Para outros erros, retorna 500 Internal Server Error
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Passa a requisição para o próximo handler
		next.ServeHTTP(w, r)
	})
}

// getClientIP extrai o IP real do cliente, considerando cabeçalhos de proxy
func getClientIP(r *http.Request) string {
	// Tenta primeiro o cabeçalho X-Forwarded-For
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// O primeiro IP na lista é o IP do cliente
		ips := net.ParseIP(ip)
		if ips != nil {
			return ips.String()
		}
	}

	// Tenta outros cabeçalhos comuns
	headers := []string{
		"X-Real-IP",
		"X-Client-IP",
	}

	for _, header := range headers {
		ip := r.Header.Get(header)
		if ip != "" {
			return ip
		}
	}

	// Se nenhum cabeçalho presente, usa o IP da conexão
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
