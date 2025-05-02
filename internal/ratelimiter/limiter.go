package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/pperesbr/ratelimiter/config"
	"github.com/pperesbr/ratelimiter/internal/ratelimiter/store"
)

// RateLimiter controla a limitação de requisições
type RateLimiter struct {
	config *LimiterConfig
	store  store.RateLimiterStore
}

// NewRateLimiter cria uma nova instância do RateLimiter
func NewRateLimiter(cfg *config.Config, storeFactory store.Factory) (*RateLimiter, error) {
	limiterStore, err := storeFactory()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar armazenamento: %w", err)
	}

	limiterConfig := &LimiterConfig{
		IPLimit:        cfg.RateLimitIP,
		IPBlockTime:    cfg.RateLimitIPBlockTime,
		TokenLimit:     cfg.RateLimitToken,
		TokenBlockTime: cfg.RateLimitTokenBlockTime,
	}

	return &RateLimiter{
		config: limiterConfig,
		store:  limiterStore,
	}, nil
}

// Allow verifica se uma requisição deve ser permitida ou bloqueada
func (rl *RateLimiter) Allow(ctx context.Context, req *LimiterRequest) error {
	// Verifica bloqueio e limites por Token (se fornecido)
	if req.Token != "" {
		tokenKey := fmt.Sprintf("token:%s", req.Token)

		// Verifica se o token está bloqueado
		blocked, err := rl.store.IsBlocked(ctx, tokenKey)
		if err != nil {
			return fmt.Errorf("erro ao verificar bloqueio do token: %w", err)
		}
		if blocked {
			return NewLimitExceededError(TokenLimit)
		}

		// Incrementa contador do token
		count, err := rl.store.IncrementRequestCount(ctx, tokenKey, time.Second)
		if err != nil {
			return fmt.Errorf("erro ao incrementar contador do token: %w", err)
		}

		// Verifica se excedeu o limite do token
		if count > rl.config.TokenLimit {
			// Bloqueia o token pelo tempo configurado
			if err := rl.store.Block(ctx, tokenKey, rl.config.TokenBlockTime); err != nil {
				return fmt.Errorf("erro ao bloquear token: %w", err)
			}
			return NewLimitExceededError(TokenLimit)
		}

		// Se o token é válido e não excedeu o limite, permite a requisição
		return nil
	}

	// Se não tem token, verifica por IP
	ipKey := fmt.Sprintf("ip:%s", req.IP)

	// Verifica se o IP está bloqueado
	blocked, err := rl.store.IsBlocked(ctx, ipKey)
	if err != nil {
		return fmt.Errorf("erro ao verificar bloqueio do IP: %w", err)
	}
	if blocked {
		return NewLimitExceededError(IPLimit)
	}

	// Incrementa contador do IP
	count, err := rl.store.IncrementRequestCount(ctx, ipKey, time.Second)
	if err != nil {
		return fmt.Errorf("erro ao incrementar contador do IP: %w", err)
	}

	// Verifica se excedeu o limite do IP
	if count > rl.config.IPLimit {
		// Bloqueia o IP pelo tempo configurado
		if err := rl.store.Block(ctx, ipKey, rl.config.IPBlockTime); err != nil {
			return fmt.Errorf("erro ao bloquear IP: %w", err)
		}
		return NewLimitExceededError(IPLimit)
	}

	// Se não excedeu o limite, permite a requisição
	return nil
}

// Close fecha o armazenamento do rate limiter
func (rl *RateLimiter) Close() error {
	return rl.store.Close()
}
