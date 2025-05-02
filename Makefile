.PHONY: build run test test-integration clean docker-build docker-run docker-compose-up docker-compose-down

# Variáveis
APP_NAME = ratelimiter
GO_FILES = $(shell find . -name "*.go" -type f -not -path "./vendor/*")
DOCKER_TAG = latest

# Cores
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

# Comandos Go
build:
	@echo "${COLOR_GREEN}Compilando aplicação...${COLOR_RESET}"
	@go build -o $(APP_NAME) ./cmd/server

run: build
	@echo "${COLOR_GREEN}Executando aplicação...${COLOR_RESET}"
	@./$(APP_NAME)

test:
	@echo "${COLOR_GREEN}Executando testes unitários...${COLOR_RESET}"
	@go test ./... -v

test-integration:
	@echo "${COLOR_GREEN}Executando testes de integração...${COLOR_RESET}"
	@TEST_INTEGRATION=1 go test ./internal/ratelimiter/store/... -v

clean:
	@echo "${COLOR_GREEN}Limpando artefatos...${COLOR_RESET}"
	@rm -f $(APP_NAME)
	@go clean -cache

# Comandos de formatação e lint
fmt:
	@echo "${COLOR_GREEN}Formatando código...${COLOR_RESET}"
	@gofmt -s -w $(GO_FILES)

lint:
	@echo "${COLOR_GREEN}Verificando código com golint...${COLOR_RESET}"
	@golint ./...

vet:
	@echo "${COLOR_GREEN}Verificando código com go vet...${COLOR_RESET}"
	@go vet ./...

# Comandos Docker
docker-build:
	@echo "${COLOR_GREEN}Construindo imagem Docker...${COLOR_RESET}"
	@docker build -t $(APP_NAME):$(DOCKER_TAG) .

docker-run: docker-build
	@echo "${COLOR_GREEN}Executando container Docker...${COLOR_RESET}"
	@docker run -p 8080:8080 --env-file .env $(APP_NAME):$(DOCKER_TAG)

# Comandos Docker Compose
docker-compose-up:
	@echo "${COLOR_GREEN}Iniciando aplicação com Docker Compose...${COLOR_RESET}"
	@docker-compose up -d

docker-compose-down:
	@echo "${COLOR_GREEN}Parando aplicação com Docker Compose...${COLOR_RESET}"
	@docker-compose down

docker-compose-logs:
	@echo "${COLOR_GREEN}Exibindo logs da aplicação...${COLOR_RESET}"
	@docker-compose logs -f

# Testes funcionais
test-functional:
	@echo "${COLOR_GREEN}Executando testes funcionais...${COLOR_RESET}"
	@chmod +x ./scripts/test.sh
	@./scripts/test.sh

load-test:
	@echo "${COLOR_GREEN}Executando teste de carga...${COLOR_RESET}"
	@chmod +x ./scripts/load_test.sh
	@./scripts/load_test.sh

# Helper
help:
	@echo "${COLOR_YELLOW}Comandos disponíveis:${COLOR_RESET}"
	@echo "  make build              - Compila a aplicação"
	@echo "  make run                - Compila e executa a aplicação"
	@echo "  make test               - Executa testes unitários"
	@echo "  make test-integration   - Executa testes de integração"
	@echo "  make clean              - Remove artefatos de compilação"
	@echo "  make fmt                - Formata o código"
	@echo "  make lint               - Executa golint"
	@echo "  make vet                - Executa go vet"
	@echo "  make docker-build       - Constrói imagem Docker"
	@echo "  make docker-run         - Executa a aplicação em um container Docker"
	@echo "  make docker-compose-up  - Inicia a aplicação com Docker Compose"
	@echo "  make docker-compose-down - Para a aplicação com Docker Compose"
	@echo "  make docker-compose-logs - Exibe logs da aplicação com Docker Compose"
	@echo "  make test-functional    - Executa testes funcionais"
	@echo "  make load-test          - Executa teste de carga"
	@echo "  make help               - Exibe esta ajuda"