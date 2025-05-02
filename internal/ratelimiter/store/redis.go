package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pperesbr/ratelimiter/config"
)

// RedisStore implementa RateLimiterStore usando Redis
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore cria uma nova instância de RedisStore
func NewRedisStore(cfg *config.Config) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Testa a conexão
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar com Redis: %w", err)
	}

	return &RedisStore{
		client: client,
	}, nil
}

// GetRequestCount obtém o número atual de requisições para uma chave
func (s *RedisStore) GetRequestCount(ctx context.Context, key string) (int, error) {
	countKey := fmt.Sprintf("count:%s", key)
	val, err := s.client.Get(ctx, countKey).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// IncrementRequestCount incrementa o contador de requisições para uma chave
func (s *RedisStore) IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int, error) {
	countKey := fmt.Sprintf("count:%s", key)

	// Incrementa o contador
	val, err := s.client.Incr(ctx, countKey).Result()
	if err != nil {
		return 0, err
	}

	// Se é a primeira requisição, define o tempo de expiração
	if val == 1 {
		s.client.Expire(ctx, countKey, expiration)
	}

	return int(val), nil
}

// IsBlocked verifica se uma chave está bloqueada
func (s *RedisStore) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := fmt.Sprintf("blocked:%s", key)
	exists, err := s.client.Exists(ctx, blockKey).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// Block bloqueia uma chave pelo tempo de bloqueio especificado
func (s *RedisStore) Block(ctx context.Context, key string, blockTime time.Duration) error {
	blockKey := fmt.Sprintf("blocked:%s", key)
	_, err := s.client.Set(ctx, blockKey, "1", blockTime).Result()
	if err != nil {
		return err
	}

	// Resetamos o contador quando bloqueamos
	countKey := fmt.Sprintf("count:%s", key)
	_, err = s.client.Del(ctx, countKey).Result()
	return err
}

// Close fecha a conexão com o Redis
func (s *RedisStore) Close() error {
	return s.client.Close()
}
