# Rate Limiter em Go

Este projeto implementa um rate limiter flexível em Go que pode ser usado como middleware em aplicações web. O limiter pode restringir requisições com base em endereço IP ou token de acesso, com diferentes limites e tempos de bloqueio configuráveis.

## Características

- Limitação por IP ou token de acesso
- Configuração flexível via variáveis de ambiente ou arquivo .env
- Implementação de Strategy Pattern para permitir diferentes mecanismos de armazenamento
- Armazenamento padrão usando Redis
- Implementação alternativa em memória para testes ou ambientes de desenvolvimento
- Alta performance e thread-safe
- Middleware HTTP facilmente integrável
- Suporte para Docker/Docker Compose

## Requisitos

- Go 1.21+
- Redis (para o storage padrão)
- Docker e Docker Compose (opcional, para execução em contêineres)

## Instalação

### Com Docker Compose

A maneira mais simples de rodar a aplicação é usando Docker Compose:

```bash
# Clone o repositório
git clone https://github.com/yourusername/ratelimiter.git
cd ratelimiter

# Inicie a aplicação com Docker Compose
docker-compose up -d
```

### Manual

Se preferir executar manualmente:

```bash
# Clone o repositório
git clone https://github.com/yourusername/ratelimiter.git
cd ratelimiter

# Baixe as dependências
go mod download

# Compile a aplicação
go build -o ratelimiter ./cmd/server

# Execute (certifique-se de ter o Redis rodando)
./ratelimiter
```

## Configuração

O rate limiter pode ser configurado através de variáveis de ambiente ou um arquivo `.env` na raiz do projeto. As seguintes variáveis estão disponíveis:

| Variável | Descrição | Valor Padrão |
|----------|-----------|--------------|
| `SERVER_PORT` | Porta do servidor HTTP | 8080 |
| `RATE_LIMIT_IP` | Requisições máximas por segundo por IP | 5 |
| `RATE_LIMIT_IP_BLOCK_TIME` | Tempo de bloqueio do IP em minutos | 5 |
| `RATE_LIMIT_TOKEN` | Requisições máximas por segundo por token | 10 |
| `RATE_LIMIT_TOKEN_BLOCK_TIME` | Tempo de bloqueio do token em minutos | 5 |
| `REDIS_HOST` | Host do Redis | localhost |
| `REDIS_PORT` | Porta do Redis | 6379 |
| `REDIS_PASSWORD` | Senha do Redis (opcional) | |
| `REDIS_DB` | Número do banco de dados Redis | 0 |
| `STORAGE_TYPE` | Tipo de armazenamento (redis ou memory) | redis |

## Como Funciona

### Middleware HTTP

O rate limiter funciona como um middleware HTTP que intercepta as requisições antes que cheguem aos handlers da aplicação. O processo de limitação ocorre da seguinte forma:

1. O middleware extrai o endereço IP do cliente e o token de acesso (se presente no header `API_KEY`).
2. Verifica se o IP ou token estão bloqueados no armazenamento.
3. Se não estiverem bloqueados, incrementa o contador de requisições.
4. Se o contador exceder o limite configurado, bloqueia o IP ou token pelo tempo definido.
5. Se o limite for excedido, retorna um erro 429 (Too Many Requests) com a mensagem apropriada.
6. Se estiver dentro do limite, permite que a requisição continue para o próximo handler.

### Prioridade Token vs IP

Quando um token de acesso é fornecido, o rate limiter prioriza as configurações do token sobre as do IP. Isso permite:

- Diferentes limites para usuários autenticados vs. não autenticados
- Limites maiores para usuários premium ou APIs parceiras
- Isolamento dos limites por usuário, independentemente do IP de origem

### Estratégia de Armazenamento

O rate limiter utiliza o padrão Strategy para permitir diferentes implementações de armazenamento:

1. **Redis (Padrão)**: Armazenamento distribuído, adequado para ambientes de produção e clusters.
2. **Memory**: Armazenamento em memória, útil para testes ou aplicações simples de um único nó.

Você pode facilmente implementar outras estratégias (como banco de dados SQL) implementando a interface `RateLimiterStore`.

## Exemplos de Uso

### Como Middleware em um Servidor HTTP

```go
package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/pperesbr/ratelimiter/config"
    "github.com/pperesbr/ratelimiter/internal/middleware"
    "github.com/pperesbr/ratelimiter/internal/ratelimiter"
    "github.com/pperesbr/ratelimiter/internal/ratelimiter/store"
)

func main() {
    // Carrega configurações
    cfg := config.LoadConfig()

    // Cria rate limiter com Redis
    storeFactory := func() (store.RateLimiterStore, error) {
        return store.NewRedisStore(cfg)
    }

    limiter, _ := ratelimiter.NewRateLimiter(cfg, storeFactory)
    defer limiter.Close()

    // Cria middleware
    rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(limiter)

    // Aplica ao router
    r := mux.NewRouter()
    r.Use(rateLimiterMiddleware.Middleware)
    
    // Adiciona handlers
    r.HandleFunc("/api/users", getUsersHandler).Methods("GET")
    
    // Inicia o servidor
    http.ListenAndServe(":"+cfg.ServerPort, r)
}
```

### Testando o Rate Limiter

Você pode testar facilmente o rate limiter usando ferramentas como `curl`:

```bash
# Requisição normal
curl http://localhost:8080/test

# Requisição com token
curl -H "API_KEY: my-token" http://localhost:8080/test

# Teste de carga com múltiplas requisições
for i in {1..10}; do curl -H "API_KEY: my-token" http://localhost:8080/test; done
```

## Testando o Projeto

Para executar os testes unitários:

```bash
go test ./...
```

Para testes de integração que requerem Redis:

```bash
# Certifique-se de que o Redis está rodando
TEST_INTEGRATION=1 go test ./internal/ratelimiter/store/...
```

# Testar requisição sem token
curl http://localhost:8080/test

# Testar requisição com token
curl -H "API_KEY: meu-token" http://localhost:8080/test

# Executar o script de teste automatizado
./scripts/test.sh

# Executar teste de carga
./scripts/load_test.sh --users 5 --requests 10 --tokens