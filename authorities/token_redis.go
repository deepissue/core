package authorities

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type redisTokenHandler struct {
	timeout time.Duration
	redis   *redis.Client
}

func NewRedisTokenHandler(redis *redis.Client, timeout time.Duration) (TokenHandler, error) {

	return &redisTokenHandler{
		redis:   redis,
		timeout: timeout,
	}, nil
}

func (r *redisTokenHandler) GenerateToken(auth *Authorized) (string, error) {

	data, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	token := uuid.New().String()
	if err := r.redis.Set(context.Background(), token, data, r.timeout).Err(); err != nil {
		return "", err
	}

	return token, nil
}

func (r *redisTokenHandler) ParseToken(token string) (*Authorized, error) {
	data, err := r.redis.Get(context.Background(), token).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var authorized Authorized
	if err := json.Unmarshal(data, &authorized); err != nil {
		return nil, err
	}

	return &authorized, nil
}
