package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/pperesbr/ratelimiter/config"
	"github.com/pperesbr/ratelimiter/internal/ratelimiter/store"
)

func TestRateLimiter_IP(t *testing.T) {
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
	limiter, err := NewRateLimiter(cfg, storeFactory)
	if err != nil {
		t.Fatalf("falha ao criar rate limiter: %v", err)
	}
	defer limiter.Close()

	// Cria contexto para testes
	ctx := context.Background()

	// Testa limite por IP
	ip := "192.168.1.1"
	req := &LimiterRequest{
		IP: ip,
	}

	// Deve permitir requisições até o limite
	for i := 0; i < cfg.RateLimitIP; i++ {
		if err := limiter.Allow(ctx, req); err != nil {
			t.Errorf("deveria permitir requisição %d, mas recebeu erro: %v", i+1, err)
		}
	}

	// A próxima requisição deve ser bloqueada
	err = limiter.Allow(ctx, req)
	if err == nil {
		t.Error("deveria bloquear requisição acima do limite, mas permitiu")
	}

	// Verifica se o erro é do tipo correto
	limitErr, ok := err.(*LimitExceededError)
	if !ok {
		t.Errorf("erro deveria ser do tipo LimitExceededError, mas recebeu: %T", err)
	}

	// Verifica se o tipo de limite é o correto
	if limitErr.Type != IPLimit {
		t.Errorf("tipo de limite deveria ser IPLimit, mas recebeu: %v", limitErr.Type)
	}
}

func TestRateLimiter_Token(t *testing.T) {
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
	limiter, err := NewRateLimiter(cfg, storeFactory)
	if err != nil {
		t.Fatalf("falha ao criar rate limiter: %v", err)
	}
	defer limiter.Close()

	// Cria contexto para testes
	ctx := context.Background()

	// Testa limite por Token
	ip := "192.168.1.1"
	token := "test-token"
	req := &LimiterRequest{
		IP:    ip,
		Token: token,
	}

	// Deve permitir requisições até o limite
	for i := 0; i < cfg.RateLimitToken; i++ {
		if err := limiter.Allow(ctx, req); err != nil {
			t.Errorf("deveria permitir requisição %d, mas recebeu erro: %v", i+1, err)
		}
	}

	// A próxima requisição deve ser bloqueada
	err = limiter.Allow(ctx, req)
	if err == nil {
		t.Error("deveria bloquear requisição acima do limite, mas permitiu")
	}

	// Verifica se o erro é do tipo correto
	limitErr, ok := err.(*LimitExceededError)
	if !ok {
		t.Errorf("erro deveria ser do tipo LimitExceededError, mas recebeu: %T", err)
	}

	// Verifica se o tipo de limite é o correto
	if limitErr.Type != TokenLimit {
		t.Errorf("tipo de limite deveria ser TokenLimit, mas recebeu: %v", limitErr.Type)
	}
}

func TestRateLimiter_TokenOverridesIP(t *testing.T) {
	// Cria configuração de teste com limite de token maior que o de IP
	cfg := &config.Config{
		RateLimitIP:             3,
		RateLimitIPBlockTime:    5 * time.Minute,
		RateLimitToken:          10, // Maior que o limite de IP
		RateLimitTokenBlockTime: 5 * time.Minute,
	}

	// Factory para armazenamento em memória
	storeFactory := func() (store.RateLimiterStore, error) {
		return store.NewMemoryStore(), nil
	}

	// Cria rate limiter
	limiter, err := NewRateLimiter(cfg, storeFactory)
	if err != nil {
		t.Fatalf("falha ao criar rate limiter: %v", err)
	}
	defer limiter.Close()

	// Cria contexto para testes
	ctx := context.Background()

	// Usa o mesmo IP para ambos os testes
	ip := "192.168.1.1"

	// Primeiro, faz requisições só com IP até quase atingir o limite
	ipReq := &LimiterRequest{
		IP: ip,
	}

	// Faz requisições até quase atingir o limite de IP
	for i := 0; i < cfg.RateLimitIP-1; i++ {
		if err := limiter.Allow(ctx, ipReq); err != nil {
			t.Errorf("deveria permitir requisição %d com IP, mas recebeu erro: %v", i+1, err)
		}
	}

	// Agora, tenta com o mesmo IP mas com token
	token := "test-token"
	tokenReq := &LimiterRequest{
		IP:    ip,
		Token: token,
	}

	// Deve permitir requisições até o limite do token, ignorando o limite do IP
	for i := 0; i < cfg.RateLimitToken; i++ {
		if err := limiter.Allow(ctx, tokenReq); err != nil {
			t.Errorf("deveria permitir requisição %d com token, mas recebeu erro: %v", i+1, err)
		}
	}

	// A próxima requisição com token deve ser bloqueada
	err = limiter.Allow(ctx, tokenReq)
	if err == nil {
		t.Error("deveria bloquear requisição acima do limite de token, mas permitiu")
	}

	// Verifica se o erro é do tipo correto e se o tipo de limite é TokenLimit
	limitErr, ok := err.(*LimitExceededError)
	if !ok || limitErr.Type != TokenLimit {
		t.Errorf("erro deveria ser LimitExceededError do tipo TokenLimit")
	}
}
