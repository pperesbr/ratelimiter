package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config contém todas as configurações da aplicação
type Config struct {
	ServerPort              string
	RateLimitIP             int
	RateLimitIPBlockTime    time.Duration
	RateLimitToken          int
	RateLimitTokenBlockTime time.Duration
	RedisHost               string
	RedisPort               string
	RedisPassword           string
	RedisDB                 int
	StorageType             string
}

// LoadConfig carrega a configuração do arquivo .env ou variáveis de ambiente
func LoadConfig() *Config {
	// Tenta carregar o arquivo .env, mas não falha se ele não existir
	err := godotenv.Load()
	if err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente")
	}

	// Carrega os valores com fallbacks para valores padrão
	rateLimitIP, _ := strconv.Atoi(getEnv("RATE_LIMIT_IP", "5"))
	rateLimitIPBlockTime, _ := strconv.Atoi(getEnv("RATE_LIMIT_IP_BLOCK_TIME", "5"))
	rateLimitToken, _ := strconv.Atoi(getEnv("RATE_LIMIT_TOKEN", "10"))
	rateLimitTokenBlockTime, _ := strconv.Atoi(getEnv("RATE_LIMIT_TOKEN_BLOCK_TIME", "5"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		ServerPort:              getEnv("SERVER_PORT", "8080"),
		RateLimitIP:             rateLimitIP,
		RateLimitIPBlockTime:    time.Duration(rateLimitIPBlockTime) * time.Minute,
		RateLimitToken:          rateLimitToken,
		RateLimitTokenBlockTime: time.Duration(rateLimitTokenBlockTime) * time.Minute,
		RedisHost:               getEnv("REDIS_HOST", "localhost"),
		RedisPort:               getEnv("REDIS_PORT", "6379"),
		RedisPassword:           getEnv("REDIS_PASSWORD", ""),
		RedisDB:                 redisDB,
		StorageType:             getEnv("STORAGE_TYPE", "redis"),
	}
}

// getEnv obtém uma variável de ambiente ou retorna um valor padrão
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
