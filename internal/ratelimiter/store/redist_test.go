package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pperesbr/ratelimiter/config"
)

// TestRedisStore_Integration é um teste de integração que requer uma instância Redis em execução
// Este teste será ignorado se a variável de ambiente TEST_INTEGRATION não estiver definida
func TestRedisStore_Integration(t *testing.T) {
	// Verifica se testes de integração estão habilitados
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("Ignorando teste de integração. Defina TEST_INTEGRATION=1 para executá-lo.")
	}

	// Configura conexão Redis
	cfg := &config.Config{
		RedisHost:     getEnvOrDefault("REDIS_HOST", "localhost"),
		RedisPort:     getEnvOrDefault("REDIS_PORT", "6379"),
		RedisPassword: getEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:       0,
	}

	// Cria store Redis
	redisStore, err := NewRedisStore(cfg)
	if err != nil {
		t.Fatalf("falha ao criar Redis store: %v", err)
	}
	defer redisStore.Close()

	// Contexto para testes
	ctx := context.Background()

	// Testa operações básicas
	t.Run("Basic Operations", func(t *testing.T) {
		// Chave de teste
		testKey := "test:integration:" + time.Now().Format(time.RFC3339)

		// Verifica contagem inicial
		count, err := redisStore.GetRequestCount(ctx, testKey)
		if err != nil {
			t.Fatalf("erro ao obter contagem: %v", err)
		}
		if count != 0 {
			t.Errorf("contagem inicial deveria ser 0, mas recebeu %d", count)
		}

		// Incrementa contador
		count, err = redisStore.IncrementRequestCount(ctx, testKey, time.Second*5)
		if err != nil {
			t.Fatalf("erro ao incrementar contador: %v", err)
		}
		if count != 1 {
			t.Errorf("contagem após incremento deveria ser 1, mas recebeu %d", count)
		}

		// Verifica novamente a contagem
		count, err = redisStore.GetRequestCount(ctx, testKey)
		if err != nil {
			t.Fatalf("erro ao obter contagem após incremento: %v", err)
		}
		if count != 1 {
			t.Errorf("contagem obtida deveria ser 1, mas recebeu %d", count)
		}
	})

	t.Run("Block/Unblock", func(t *testing.T) {
		// Chave de teste
		testKey := "test:block:" + time.Now().Format(time.RFC3339)

		// Verifica se não está bloqueado inicialmente
		blocked, err := redisStore.IsBlocked(ctx, testKey)
		if err != nil {
			t.Fatalf("erro ao verificar bloqueio: %v", err)
		}
		if blocked {
			t.Error("chave não deveria estar bloqueada inicialmente")
		}

		// Bloqueia a chave por um curto período
		blockTime := 2 * time.Second
		err = redisStore.Block(ctx, testKey, blockTime)
		if err != nil {
			t.Fatalf("erro ao bloquear chave: %v", err)
		}

		// Verifica se está bloqueado
		blocked, err = redisStore.IsBlocked(ctx, testKey)
		if err != nil {
			t.Fatalf("erro ao verificar bloqueio após block: %v", err)
		}
		if !blocked {
			t.Error("chave deveria estar bloqueada após Block()")
		}

		// Espera o bloqueio expirar
		time.Sleep(blockTime + time.Second)

		// Verifica se o bloqueio expirou
		blocked, err = redisStore.IsBlocked(ctx, testKey)
		if err != nil {
			t.Fatalf("erro ao verificar bloqueio após expiração: %v", err)
		}
		if blocked {
			t.Error("chave não deveria estar bloqueada após expiração")
		}
	})

	t.Run("Expiration", func(t *testing.T) {
		// Chave de teste
		testKey := "test:expiry:" + time.Now().Format(time.RFC3339)

		// Incrementa com expiração curta
		expiration := 2 * time.Second
		_, err := redisStore.IncrementRequestCount(ctx, testKey, expiration)
		if err != nil {
			t.Fatalf("erro ao incrementar contador com expiração: %v", err)
		}

		// Verifica se o contador existe
		count, err := redisStore.GetRequestCount(ctx, testKey)
		if err != nil {
			t.Fatalf("erro ao obter contagem após incremento: %v", err)
		}
		if count != 1 {
			t.Errorf("contagem deveria ser 1, mas recebeu %d", count)
		}

		// Espera a expiração
		time.Sleep(expiration + time.Second)

		// Verifica se o contador expirou
		count, err = redisStore.GetRequestCount(ctx, testKey)
		if err != nil {
			t.Fatalf("erro ao obter contagem após expiração: %v", err)
		}
		if count != 0 {
			t.Errorf("contagem deveria ser 0 após expiração, mas recebeu %d", count)
		}
	})
}

// Utilitário para obter variável de ambiente com valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
