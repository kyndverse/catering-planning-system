package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func NewRedisClient(config *ConfigEnv, logger *logrus.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(
			"%s:%d",
			config.REDIS_HOST,
			config.REDIS_PORT,
		),
		Password: config.REDIS_PASSWORD,
		DB:       config.REDIS_DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Fatalf("failed to connect to redis: %v", err)
	}

	logger.Info("Redis connected")

	return client
}
