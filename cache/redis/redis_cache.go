package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRemoteCache struct {
	client *redis.Client
}

func NewRedisRemoteCache(client *redis.Client) *RedisRemoteCache {
	return &RedisRemoteCache{
		client: client,
	}
}

func (c *RedisRemoteCache) PutToRemoteCache(ctx context.Context, key string, value any, timeToLiveInSeconds ...int) {
	var duration time.Duration
	if len(timeToLiveInSeconds) > 0 {
		duration = time.Duration(timeToLiveInSeconds[0]) * time.Second
	}
	
	bytes, err := json.Marshal(value)
	if err == nil {
		c.client.Set(ctx, key, string(bytes), duration)
	}
}

func (c *RedisRemoteCache) GetFromRemoteCache(ctx context.Context, key string) any {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil
	}
	var res any
	if err := json.Unmarshal([]byte(val), &res); err == nil {
		return res
	}
	return nil
}

func (c *RedisRemoteCache) RemoveFromRemoteCache(ctx context.Context, key string) {
	c.client.Del(ctx, key)
}
