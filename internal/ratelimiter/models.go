package ratelimiter

import "time"

// LimiterConfig armazena a configuração do rate limiter
type LimiterConfig struct {
	// Limite de requisições por segundo por IP
	IPLimit int
	// Tempo de bloqueio para IP excedido
	IPBlockTime time.Duration

	// Limite de requisições por segundo por Token
	TokenLimit int
	// Tempo de bloqueio para Token excedido
	TokenBlockTime time.Duration
}

// LimitType define o tipo de limitação (IP ou Token)
type LimitType int

const (
	// IPLimit representa limitação por IP
	IPLimit LimitType = iota
	// TokenLimit representa limitação por Token
	TokenLimit
)

// LimiterRequest representa uma requisição ao rate limiter
type LimiterRequest struct {
	// IP do cliente
	IP string
	// Token de acesso (opcional)
	Token string
}

// LimitExceededError é retornado quando o limite é excedido
type LimitExceededError struct {
	// Tipo de limitação que foi excedida
	Type LimitType
	// Mensagem de erro
	Message string
}

// Error implementa a interface error
func (e *LimitExceededError) Error() string {
	return e.Message
}

// NewLimitExceededError cria um novo erro de limite excedido
func NewLimitExceededError(limitType LimitType) *LimitExceededError {
	return &LimitExceededError{
		Type:    limitType,
		Message: "you have reached the maximum number of requests or actions allowed within a certain time frame",
	}
}
