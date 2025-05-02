package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response representa a resposta JSON para as requisições
type Response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewResponse cria uma nova resposta
func NewResponse(message string) Response {
	return Response{
		Message:   message,
		Timestamp: time.Now(),
	}
}

// HomeHandler retorna um handler simples para a rota raiz
func HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := NewResponse("Welcome to the Rate Limiter API")
		jsonResponse(w, response, http.StatusOK)
	}
}

// TestHandler retorna um handler para testar o rate limiter
func TestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extrai o token (se presente) e inclui na resposta para facilitar testes
		token := r.Header.Get("API_KEY")
		message := "Request successful"

		if token != "" {
			message = "Request successful with token: " + token
		}

		response := NewResponse(message)
		jsonResponse(w, response, http.StatusOK)
	}
}

// jsonResponse envia uma resposta JSON para o cliente
func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
