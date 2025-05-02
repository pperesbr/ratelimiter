FROM golang:1.24.2-alpine AS builder

# Diretório de trabalho
WORKDIR /app

# Dependências
RUN apk add --no-cache git

# Copia apenas os arquivos necessários para baixar dependências
COPY go.mod go.sum ./

# Baixa as dependências
RUN go mod download

# Copia o restante do código
COPY . .

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o ratelimiter ./cmd/server

# Imagem final
FROM alpine:3.18

# Instala dependências necessárias
RUN apk add --no-cache ca-certificates tzdata

# Define o fuso horário para UTC
ENV TZ=UTC

# Cria um usuário não-root
RUN adduser -D -h /app appuser
USER appuser

# Diretório de trabalho
WORKDIR /app

# Copia o binário compilado
COPY --from=builder /app/ratelimiter .

# Expõe a porta
EXPOSE 8080

# Comando para iniciar a aplicação
CMD ["./ratelimiter"]