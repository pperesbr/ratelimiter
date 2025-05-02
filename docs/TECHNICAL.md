# Documentação Técnica - Rate Limiter

## Visão Geral da Arquitetura

O Rate Limiter foi desenvolvido seguindo princípios de Clean Architecture e Design Patterns para garantir flexibilidade, testabilidade e manutenibilidade.

### Arquitetura

```
┌────────────────────┐      ┌───────────────────┐      ┌───────────────────┐
│                    │      │                   │      │                   │
│   HTTP Middleware  │─────▶│   Rate Limiter    │─────▶│   Storage Layer   │
│                    │      │                   │      │                   │
└────────────────────┘      └───────────────────┘      └───────────────────┘
                                                               │
                                                               │
                                                               ▼
                                                        ┌─────────────┐
                                                        │             │
                                                        │    Redis    │
                                                        │             │
                                                        └─────────────┘
```

### Principais Componentes

1. **HTTP Middleware**: Intercepta requisições HTTP, extrai informações relevantes (IP, token) e aplica as regras de limitação.
2. **Rate Limiter**: Implementa a lógica de negócio para controlar taxas de requisição.
3. **Storage Layer**: Abstração que permite diferentes implementações de armazenamento.
4. **Redis**: Implementação padrão para armazenar contadores e estados de bloqueio.

## Padrões de Design Utilizados

### Strategy Pattern

O padrão Strategy foi utilizado para permitir a troca fácil do mecanismo de armazenamento:

```
┌─────────────────┐
│                 │
│  RateLimiter    │
│                 │
└───────┬─────────┘
        │
        │ usa
        ▼
┌─────────────────┐     ┌────────────────┐
│                 │     │                │
│ RateLimiterStore│◄────┤ RedisStore     │
│   (interface)   │     │                │
│                 │     └────────────────┘
└─────────────────┘
        ▲
        │
        │
┌───────┴─────────┐
│                 │
│  MemoryStore    │
│                 │
└─────────────────┘
```

Este padrão permite que novas implementações de armazenamento sejam adicionadas facilmente sem modificar o código do Rate Limiter.

### Dependency Injection

Utilizamos injeção de dependência para fornecer a implementação de armazenamento ao Rate Limiter, facilitando os testes e a configuração:

```go
// Factory para criação do armazenamento
storeFactory := func() (store.RateLimiterStore, error) {
    return store.NewRedisStore(cfg)
}

// Injeção da dependência no Rate Limiter
limiter, err := ratelimiter.NewRateLimiter(cfg, storeFactory)
```

### Middleware Pattern

O padrão Middleware foi utilizado para integrar o Rate Limiter com o servidor HTTP:

```go
// Aplica o middleware ao router
r := mux.NewRouter()
r.Use(rateLimiterMiddleware.Middleware)
```

## Detalhes de Implementação

### Modelo de Dados

#### LimiterConfig

```go
type LimiterConfig struct {
    IPLimit        int
    IPBlockTime    time.Duration
    TokenLimit     int
    TokenBlockTime time.Duration
}
```

Armazena as configurações de limite de requisições e tempo de bloqueio.

#### LimiterRequest

```go
type LimiterRequest struct {
    IP    string
    Token string
}
```

Representa uma requisição ao Rate Limiter, contendo o IP e o token (opcional).

### Interface de Armazenamento

```go
type RateLimiterStore interface {
    GetRequestCount(ctx context.Context, key string) (int, error)
    IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int, error)
    IsBlocked(ctx context.Context, key string) (bool, error)
    Block(ctx context.Context, key string, blockTime time.Duration) error
    Close() error
}
```

Esta interface define as operações necessárias para qualquer implementação de armazenamento:

- `GetRequestCount`: Obtém o número atual de requisições para uma chave.
- `IncrementRequestCount`: Incrementa o contador de requisições.
- `IsBlocked`: Verifica se uma chave está bloqueada.
- `Block`: Bloqueia uma chave pelo tempo especificado.
- `Close`: Fecha a conexão com o armazenamento.

### Implementações de Armazenamento

#### RedisStore

Implementação de armazenamento usando Redis, adequada para ambientes de produção:

- Usa `INCR` para contadores atômicos.
- Usa `SET` com expiração para bloqueios.
- Suporta clusters Redis.

#### MemoryStore

Implementação em memória para testes ou desenvolvimento:

- Armazena contadores e bloqueios em maps.
- Thread-safe usando mutex.
- Inclui rotina de limpeza para evitar vazamento de memória.

### Fluxo de Processamento

1. **Recebimento da Requisição**: O middleware intercepta a requisição HTTP.
2. **Extração de Dados**: O IP do cliente e o token (se presente) são extraídos.
3. **Verificação de Bloqueio**: Verifica se o IP ou token estão bloqueados.
4. **Verificação de Limite**:
    - Se um token está presente, verifica o limite do token.
    - Caso contrário, verifica o limite do IP.
5. **Incremento do Contador**: Incrementa o contador de requisições.
6. **Aplicação de Bloqueio**: Se o limite for excedido, o IP ou token é bloqueado pelo tempo configurado.
7. **Resposta**:
    - Se bloqueado ou limite excedido: Retorna 429 Too Many Requests.
    - Caso contrário: Passa a requisição para o próximo handler.

## Considerações de Performance

### Redis

- Operações são O(1) e muito rápidas.
- Recomendado usar Redis em modo cluster para alta disponibilidade.
- Configurar `maxmemory` e política de evicção para evitar problemas de memória.

### Memória

- Todas as operações são O(1).
- Utiliza rotina de limpeza para evitar vazamento de memória.
- Não recomendado para ambientes de produção com múltiplas instâncias.

## Extensões Possíveis

### Outros Armazenamentos

Você pode implementar outras estratégias de armazenamento:

- **Database**: Armazenamento em banco de dados SQL/NoSQL.
- **Distributed Cache**: Outros sistemas de cache como Memcached.
- **Hybrid**: Combinação de diferentes mecanismos para diferentes tipos de limitação.

### Limites Dinâmicos

O sistema pode ser estendido para suportar limites dinâmicos baseados em:

- Hora do dia
- Carga do sistema
- Tipo de usuário (planos diferentes)
- Comportamento do usuário

### Monitoramento

Implementar métricas para monitoramento:

- Taxa de bloqueio
- Padrões de uso
- Detecção de abusos

## Resolução de Problemas

### Limites muito restritivos

Se os limites estiverem muito restritivos:

1. Aumente os valores de `RATE_LIMIT_IP` e `RATE_LIMIT_TOKEN` no arquivo `.env`.
2. Reinicie a aplicação.

### Problemas de Conectividade com Redis

Se houver problemas de conexão com Redis:

1. Verifique se o Redis está rodando: `docker-compose ps`.
2. Verifique as configurações de host/porta no arquivo `.env`.
3. Teste a conexão diretamente: `redis-cli ping`.

### Falsos Positivos (Cliente Bloqueado Incorretamente)

Possíveis causas:

1. Múltiplos usuários compartilhando o mesmo IP (NAT/proxy).
2. Problemas de sincronização em ambiente de cluster.

Soluções:

1. Use tokens de autenticação para usuários legítimos.
2. Ajuste os limites para acomodar múltiplos usuários por IP.
3. Implemente whitelist para IPs confiáveis.