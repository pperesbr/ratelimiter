#!/bin/bash

# Script para teste de carga do rate limiter

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# URL base da API
BASE_URL="http://localhost:8080/test"

# Configurações
CONCURRENT_USERS=10
REQUESTS_PER_USER=20
SLEEP_BETWEEN_REQUESTS=0.1
USE_TOKENS=false

# Função para exibir ajuda
show_help() {
  echo -e "${BLUE}Teste de Carga para Rate Limiter${NC}"
  echo ""
  echo "Uso: $0 [OPÇÕES]"
  echo ""
  echo "Opções:"
  echo "  -u, --users NUM       Número de usuários concorrentes (padrão: 10)"
  echo "  -r, --requests NUM    Requisições por usuário (padrão: 20)"
  echo "  -s, --sleep NUM       Tempo de espera entre requisições em segundos (padrão: 0.1)"
  echo "  -t, --tokens          Usar tokens únicos para cada usuário"
  echo "  -h, --help            Exibe esta ajuda"
  echo ""
  echo "Exemplo:"
  echo "  $0 --users 5 --requests 10 --tokens"
  echo ""
}

# Processar argumentos
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    -u|--users)
      CONCURRENT_USERS="$2"
      shift
      shift
      ;;
    -r|--requests)
      REQUESTS_PER_USER="$2"
      shift
      shift
      ;;
    -s|--sleep)
      SLEEP_BETWEEN_REQUESTS="$2"
      shift
      shift
      ;;
    -t|--tokens)
      USE_TOKENS=true
      shift
      ;;
    -h|--help)
      show_help
      exit 0
      ;;
    *)
      echo -e "${RED}Opção desconhecida: $1${NC}"
      show_help
      exit 1
      ;;
  esac
done

# Exibe configurações do teste
echo -e "${BLUE}=== Configuração do Teste ===${NC}"
echo -e "Usuários concorrentes: ${YELLOW}$CONCURRENT_USERS${NC}"
echo -e "Requisições por usuário: ${YELLOW}$REQUESTS_PER_USER${NC}"
echo -e "Tempo entre requisições: ${YELLOW}${SLEEP_BETWEEN_REQUESTS}s${NC}"
echo -e "Usar tokens únicos: ${YELLOW}$USE_TOKENS${NC}"
echo ""
echo -e "${BLUE}=== Iniciando Teste ===${NC}"
echo ""

# Variáveis para estatísticas
total_requests=0
successful_requests=0
blocked_requests=0
start_time=$(date +%s.%N)

# Função para executar requisições de um usuário
run_user_requests() {
  local user_id=$1
  local token=""
  local header=""
  local total_user_requests=0
  local successful_user_requests=0
  local blocked_user_requests=0

  # Gera um token único se necessário
  if [ "$USE_TOKENS" = true ]; then
    token="test-token-$user_id"
    header="API_KEY: $token"
  fi

  echo -e "${YELLOW}Usuário $user_id iniciando${NC}"

  for ((i=1; i<=$REQUESTS_PER_USER; i++)); do
    # Atualiza contadores
    total_user_requests=$((total_user_requests + 1))

    # Faz a requisição
    if [ -z "$header" ]; then
      response=$(curl -s -w "\n%{http_code}" "$BASE_URL")
    else
      response=$(curl -s -w "\n%{http_code}" -H "$header" "$BASE_URL")
    fi

    # Extrai o status code
    status_code=$(echo "$response" | tail -n1)

    # Atualiza contadores baseado no resultado
    if [ "$status_code" -eq 200 ]; then
      successful_user_requests=$((successful_user_requests + 1))
    else
      blocked_user_requests=$((blocked_user_requests + 1))
    fi

    # Espera um pouco antes da próxima requisição
    sleep $SLEEP_BETWEEN_REQUESTS
  done

  # Retorna estatísticas do usuário
  echo "$total_user_requests $successful_user_requests $blocked_user_requests"
}

# Executa os testes em paralelo
for ((user=1; user<=$CONCURRENT_USERS; user++)); do
  run_user_requests $user &
done

# Aguarda a conclusão de todos os processos
wait

# Calcula o tempo total
end_time=$(date +%s.%N)
total_time=$(echo "$end_time - $start_time" | bc)

# Exibe estatísticas finais
echo ""
echo -e "${BLUE}=== Estatísticas do Teste ===${NC}"
echo -e "Tempo total de execução: ${YELLOW}$(printf "%.2f" $total_time)s${NC}"
echo -e "Requisições totais: ${YELLOW}$((CONCURRENT_USERS * REQUESTS_PER_USER))${NC}"
echo -e "Taxa de requisições: ${YELLOW}$(printf "%.2f" $(echo "$CONCURRENT_USERS * $REQUESTS_PER_USER / $total_time" | bc -l))/s${NC}"
echo ""
echo -e "${GREEN}Teste de carga concluído.${NC}"