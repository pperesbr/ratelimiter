package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pperesbr/ratelimiter/config"
	"github.com/pperesbr/ratelimiter/internal/handlers"
	"github.com/pperesbr/ratelimiter/internal/middleware"
	"github.com/pperesbr/ratelimiter/internal/ratelimiter"
	"github.com/pperesbr/ratelimiter/internal/ratelimiter/store"
)

func main() {
	// Carrega configurações
	cfg := config.LoadConfig()

	// Cria factory de armazenamento baseado no tipo configurado
	var storeFactory store.Factory
	switch cfg.StorageType {
	case "redis":
		storeFactory = func() (store.RateLimiterStore, error) {
			return store.NewRedisStore(cfg)
		}
	case "memory":
		storeFactory = func() (store.RateLimiterStore, error) {
			return store.NewMemoryStore(), nil
		}
	default:
		log.Fatalf("Tipo de armazenamento não suportado: %s", cfg.StorageType)
	}

	// Cria rate limiter
	limiter, err := ratelimiter.NewRateLimiter(cfg, storeFactory)
	if err != nil {
		log.Fatalf("Falha ao criar rate limiter: %v", err)
	}
	defer limiter.Close()

	// Cria middleware do rate limiter
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(limiter)

	// Cria router e define rotas
	r := mux.NewRouter()
	r.Use(rateLimiterMiddleware.Middleware)

	// Define os handlers
	r.HandleFunc("/", handlers.HomeHandler()).Methods("GET")
	r.HandleFunc("/test", handlers.TestHandler()).Methods("GET")

	// Configuração do servidor
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Inicia o servidor em uma goroutine separada
	go func() {
		log.Printf("Servidor iniciado na porta %s\n", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Falha ao iniciar servidor: %v", err)
		}
	}()

	// Configura graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Bloqueia até receber um sinal de interrupção
	<-quit
	log.Println("Servidor está sendo encerrado...")

	// Cria um contexto com timeout para o shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Encerra o servidor
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Falha ao encerrar servidor: %v", err)
	}

	log.Println("Servidor encerrado com sucesso")
}
