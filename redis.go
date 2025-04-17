package tgstatemanager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisStorage provides Redis-backed storage for user states.
type RedisStorage[S any] struct {
	client *redis.Client
	ctx    context.Context
	prefix string
}

// NewRedisStorage creates a new Redis storage instance.
func NewRedisStorage[S any](client *redis.Client, prefix string) *RedisStorage[S] {
	return &RedisStorage[S]{
		client: client,
		ctx:    context.Background(),
		prefix: prefix,
	}
}

// formatKey creates a consistent Redis key for a user ID.
func (s *RedisStorage[S]) formatKey(id int64) string {
	return fmt.Sprintf("%s:%d", s.prefix, id)
}

// Get retrieves a user state from Redis.
func (s *RedisStorage[S]) Get(id int64) (UserState[S], bool, error) {
	data, err := s.client.Get(s.ctx, s.formatKey(id)).Bytes()

	// Handle non-existent key
	if err == redis.Nil {
		return UserState[S]{}, false, nil
	}

	// Handle other Redis errors
	if err != nil {
		return UserState[S]{}, false, err
	}

	// Unmarshal data
	var state UserState[S]
	if err := json.Unmarshal(data, &state); err != nil {
		return UserState[S]{}, false, err
	}

	return state, true, nil
}

// Set stores a user state in Redis.
func (s *RedisStorage[S]) Set(id int64, state UserState[S]) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return s.client.Set(s.ctx, s.formatKey(id), data, 0).Err()
}
