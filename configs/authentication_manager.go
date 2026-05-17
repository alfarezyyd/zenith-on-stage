package configs

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisAuthManager struct {
	redisClient *redis.Client
	PrefixState string
	timeToLive  time.Duration
}

func NewRedisAuthManager(redisClient *redis.Client) *RedisAuthManager {
	return &RedisAuthManager{
		redisClient: redisClient,
		PrefixState: "stateauth",
		timeToLive:  2 * time.Minute,
	}
}
func (redisAuthManager *RedisAuthManager) buildKeyState(state string) string {
	return fmt.Sprintf("%s:%s", redisAuthManager.PrefixState, state)
}

func (redisAuthManager *RedisAuthManager) SetState(ctx context.Context, state string) error {
	key := redisAuthManager.buildKeyState(state)
	expiration := redisAuthManager.timeToLive

	err := redisAuthManager.redisClient.Set(ctx, key, state, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}
	return nil
}

func (redisAuthManager *RedisAuthManager) GetState(
	ctx context.Context,
	state string,
) (string, error) {
	key := redisAuthManager.buildKeyState(state)
	stateData, err := redisAuthManager.redisClient.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get state data from Redis: %w", err)
	}
	return stateData, nil
}
func (redisAuthManager *RedisAuthManager) DeleteState(
	ctx context.Context,
	state string,
) error {
	key := redisAuthManager.buildKeyState(state)
	if err := redisAuthManager.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove state data from Redis: %w", err)
	}
	return nil
}
