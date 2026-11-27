package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-web/internal/domain/entity"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	redis *redis.Client
}

func NewSessionRepository(redisClient *redis.Client) *SessionRepository {
	return &SessionRepository{
		redis: redisClient,
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *entity.Session, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := sessionKey(session.ID)

	return r.redis.Set(ctx, key, data, ttl).Err()
}

func (r *SessionRepository) FindByID(ctx context.Context, sessionID string) (*entity.Session, error) {
	key := sessionKey(sessionID)

	value, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var session entity.Session
	if err := json.Unmarshal(value, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	return r.redis.Del(ctx, sessionKey(sessionID)).Err()
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
