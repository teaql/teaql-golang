package redis

import (
	stdcontext "context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRemoteLock struct {
	client *redis.Client
}

func NewRedisRemoteLock(client *redis.Client) *RedisRemoteLock {
	return &RedisRemoteLock{
		client: client,
	}
}

func (l *RedisRemoteLock) TryRemoteLock(context stdcontext.Context, key string, timeoutMillis int64, expireMillis int64) bool {
	deadline := time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	expire := time.Duration(expireMillis) * time.Millisecond

	for {
		ok, err := l.client.SetNX(context, key, "locked", expire).Result()
		if err == nil && ok {
			return true
		}

		if time.Now().After(deadline) {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func (l *RedisRemoteLock) UnlockRemote(context stdcontext.Context, key string) {
	l.client.Del(context, key)
}
