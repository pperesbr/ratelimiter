package store

import (
	"context"
	"time"
)

// RateLimiterStore define a interface para armazenamento do rate limiter
type RateLimiterStore interface {
	// GetRequestCount obtém o número atual de requisições para uma chave
	GetRequestCount(ctx context.Context, key string) (int, error)

	// IncrementRequestCount incrementa o contador de requisições para uma chave
	IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int, error)

	// IsBlocked verifica se uma chave está bloqueada
	IsBlocked(ctx context.Context, key string) (bool, error)

	// Block bloqueia uma chave pelo tempo de bloqueio especificado
	Block(ctx context.Context, key string, blockTime time.Duration) error

	// Close fecha a conexão com o armazenamento
	Close() error
}

// Factory define um tipo para funções factory de RateLimiterStore
type Factory func() (RateLimiterStore, error)
