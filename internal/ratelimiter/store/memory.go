package store

import (
	"context"
	"sync"
	"time"
)

// MemoryStore implementa RateLimiterStore usando armazenamento em memória
type MemoryStore struct {
	counts     map[string]int
	blocked    map[string]time.Time
	expiryTime map[string]time.Time
	mu         sync.RWMutex
}

// NewMemoryStore cria uma nova instância de MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		counts:     make(map[string]int),
		blocked:    make(map[string]time.Time),
		expiryTime: make(map[string]time.Time),
	}
}

// GetRequestCount obtém o número atual de requisições para uma chave
func (s *MemoryStore) GetRequestCount(ctx context.Context, key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Verifica se o tempo expirou
	if expiry, ok := s.expiryTime[key]; ok && time.Now().After(expiry) {
		// Normalmente, removemos a chave expirada em IncrementRequestCount, mas aqui
		// só precisamos retornar 0 porque está expirada
		return 0, nil
	}

	return s.counts[key], nil
}

// IncrementRequestCount incrementa o contador de requisições para uma chave
func (s *MemoryStore) IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verifica se o tempo expirou
	if expiry, ok := s.expiryTime[key]; ok && time.Now().After(expiry) {
		// Limpa os dados expirados
		delete(s.counts, key)
		delete(s.expiryTime, key)
	}

	// Incrementa o contador
	s.counts[key]++

	// Se é a primeira requisição, define o tempo de expiração
	if s.counts[key] == 1 {
		s.expiryTime[key] = time.Now().Add(expiration)
	}

	return s.counts[key], nil
}

// IsBlocked verifica se uma chave está bloqueada
func (s *MemoryStore) IsBlocked(ctx context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blockUntil, blocked := s.blocked[key]
	return blocked && time.Now().Before(blockUntil), nil
}

// Block bloqueia uma chave pelo tempo de bloqueio especificado
func (s *MemoryStore) Block(ctx context.Context, key string, blockTime time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blocked[key] = time.Now().Add(blockTime)

	// Resetamos o contador quando bloqueamos
	delete(s.counts, key)
	delete(s.expiryTime, key)

	return nil
}

// Close fecha o armazenamento (nada a fazer para implementação em memória)
func (s *MemoryStore) Close() error {
	return nil
}

// Rotina de limpeza de dados expirados (não é necessária para o Redis, que tem seu próprio mecanismo de expiração)
func (s *MemoryStore) startCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.cleanup()
		}
	}()
}

// Limpa dados expirados
func (s *MemoryStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Limpa contadores expirados
	for key, expiry := range s.expiryTime {
		if now.After(expiry) {
			delete(s.counts, key)
			delete(s.expiryTime, key)
		}
	}

	// Limpa bloqueios expirados
	for key, blockUntil := range s.blocked {
		if now.After(blockUntil) {
			delete(s.blocked, key)
		}
	}
}
