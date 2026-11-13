package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ConfigEnv struct {
	ADDR      string
	DB_DSN    string
	LOG_LEVEL string
}

func NewEnvConfig() *ConfigEnv {
	_ = godotenv.Load()

	cfg := &ConfigEnv{
		ADDR:      getEnv("ADDR", ":8080"),
		DB_DSN:    os.Getenv("DB_DSN"),
		LOG_LEVEL: os.Getenv("LOG_LEVEL"),
	}

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
