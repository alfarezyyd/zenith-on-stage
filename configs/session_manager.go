package configs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"zenith-on-stage/internal/model"

	"github.com/redis/go-redis/v9"
)

type RedisSessionManager struct {
	redisClient *redis.Client
	PrefixState string
	defaultTTL  time.Duration
}

func NewSessionRedisManager(redisClient *redis.Client) *RedisSessionManager {
	return &RedisSessionManager{
		redisClient: redisClient,
		PrefixState: "session",
		defaultTTL:  2 * time.Hour,
	}
}
func (redisSessionManager *RedisSessionManager) buildKeyState(session string) string {
	return fmt.Sprintf("%s:%s", redisSessionManager.PrefixState, session)
}

// Set stores session data in Redis
func (redisSessionManager *RedisSessionManager) Set(ctx context.Context, sessionID string, data model.SessionData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := redisSessionManager.buildKeyState(sessionID)
	return redisSessionManager.redisClient.Set(ctx, key, jsonData, redisSessionManager.defaultTTL).Err()
}

// Get retrieves session data from Redis
func (redisSessionManager *RedisSessionManager) Get(ctx context.Context, sessionID string) (*model.SessionData, error) {
	key := redisSessionManager.buildKeyState(sessionID)
	data, err := redisSessionManager.redisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var sessionData model.SessionData
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &sessionData, nil
}

// Delete removes a session from Redis
func (redisSessionManager *RedisSessionManager) Delete(ctx context.Context, sessionID string) error {
	key := redisSessionManager.buildKeyState(sessionID)
	return redisSessionManager.redisClient.Del(ctx, key).Err()
}
