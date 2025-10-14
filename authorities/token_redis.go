package authorities

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type redisTokenHandler struct {
	settings *Settings
	redis    *redis.Client
}

func NewRedisTokenHandler(settings *Settings, redis *redis.Client) (TokenHandler, error) {

	return &redisTokenHandler{
		settings: settings,
		redis:    redis,
	}, nil
}

func (r *redisTokenHandler) GenerateToken(auth *Authorized) (string, error) {

	data, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	token := uuid.New().String()
	if err := r.redis.Set(context.Background(), token, data, r.settings.Timeout).Err(); err != nil {
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
