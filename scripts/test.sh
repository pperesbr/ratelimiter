#!/bin/bash

# Script para testar o rate limiter

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# URL base da API
BASE_URL="http://localhost:8080"

# Função para fazer requisição e mostrar resultado
make_request() {
    local url=$1
    local headers=$2
    local description=$3

    echo -e "${YELLOW}Teste: $description${NC}"

    # Faz a requisição
    if [ -z "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -H "$headers" "$url")
    fi

    # Extrai o status code
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    # Mostra resultado
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}✓ Sucesso (HTTP $status_code)${NC}"
    else
        echo -e "${RED}✗ Falhou (HTTP $status_code)${NC}"
    fi

    echo "Resposta: $body"
    echo ""
}

# Função para testar limite por IP
test_ip_limit() {
    echo -e "${YELLOW}===== Testando Limite por IP =====${NC}"

    # Faz várias requisições até exceder o limite
    for i in {1..7}; do
        echo -e "Requisição $i"
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL/test")

        status_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')

        if [ "$status_code" -eq 200 ]; then
            echo -e "${GREEN}✓ Sucesso (HTTP $status_code)${NC}"
        else
            echo -e "${RED}✗ Bloqueado (HTTP $status_code)${NC}"
            echo "Resposta: $body"
            break
        fi

        # Pequena pausa para não estourar muito rápido
        sleep 0.1
    done

    echo ""
}

# Função para testar limite por token
test_token_limit() {
    local token=$1

    echo -e "${YELLOW}===== Testando Limite por Token ($token) =====${NC}"

    # Faz várias requisições até exceder o limite
    for i in {1..12}; do
        echo -e "Requisição $i"
        response=$(curl -s -w "\n%{http_code}" -H "API_KEY: $token" "$BASE_URL/test")

        status_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')

        if [ "$status_code" -eq 200 ]; then
            echo -e "${GREEN}✓ Sucesso (HTTP $status_code)${NC}"
        else
            echo -e "${RED}✗ Bloqueado (HTTP $status_code)${NC}"
            echo "Resposta: $body"
            break
        fi

        # Pequena pausa para não estourar muito rápido
        sleep 0.1
    done

    echo ""
}

# Verifica se o servidor está rodando
echo -e "${YELLOW}Verificando se o servidor está rodando...${NC}"
if ! curl -s "$BASE_URL" > /dev/null; then
    echo -e "${RED}Erro: Servidor não está rodando em $BASE_URL${NC}"
    echo "Inicie o servidor com 'docker-compose up -d' ou 'go run cmd/server/main.go'"
    exit 1
fi

# Testes básicos
make_request "$BASE_URL" "" "Rota raiz (sem token)"
make_request "$BASE_URL/test" "" "Rota de teste (sem token)"
make_request "$BASE_URL/test" "API_KEY: test-token" "Rota de teste (com token)"

# Testes de limite
test_ip_limit
test_token_limit "test-token-1"

echo -e "${YELLOW}===== Testes Completos =====${NC}"