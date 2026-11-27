package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ConfigEnv struct {
	ADDR           string
	DB_DSN         string
	LOG_LEVEL      string
	REDIS_HOST     string
	REDIS_PORT     int
	REDIS_PASSWORD string
	REDIS_DB       int
}

func NewEnvConfig() *ConfigEnv {
	_ = godotenv.Load()

	cfg := &ConfigEnv{
		ADDR:           getEnv("ADDR", ":8080"),
		DB_DSN:         os.Getenv("DB_DSN"),
		LOG_LEVEL:      os.Getenv("LOG_LEVEL"),
		REDIS_HOST:     getEnv("REDIS_HOST", "localhost"),
		REDIS_PORT:     parseInt(getEnv("REDIS_PORT", "6379")),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		REDIS_DB:       parseInt(os.Getenv("REDIS_DB")),
	}

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseInt(s string) int {
	number, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return number
}
