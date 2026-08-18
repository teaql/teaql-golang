package redis

import (
	stdcontext "context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/teaql/teaql-golang/core"
)

type DataStore interface {
	Get(context stdcontext.Context, key string) (core.Value, bool)
	Put(context stdcontext.Context, key string, value core.Value, timeoutSeconds *uint64) error
	Remove(context stdcontext.Context, key string) error
}

type RedisDataStore struct {
	client *redis.Client
}

func NewRedisDataStore(redisUrl string) (*RedisDataStore, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	return &RedisDataStore{
		client: client,
	}, nil
}

func (s *RedisDataStore) Client() *redis.Client {
	return s.client
}

func (s *RedisDataStore) Get(context stdcontext.Context, key string) (core.Value, bool) {
	val, err := s.client.Get(context, key).Result()
	if err == redis.Nil || err != nil {
		return core.ValNull(), false
	}

	var jsonVal interface{}
	if err := json.Unmarshal([]byte(val), &jsonVal); err != nil {
		return core.ValNull(), false
	}

	// simple conversion
	return jsonToCoreValue(jsonVal), true
}

func (s *RedisDataStore) Put(context stdcontext.Context, key string, value core.Value, timeoutSeconds *uint64) error {
	jsonBytes, err := json.Marshal(value.V)
	if err != nil {
		return err
	}

	var duration time.Duration
	if timeoutSeconds != nil {
		duration = time.Duration(*timeoutSeconds) * time.Second
	}

	return s.client.Set(context, key, string(jsonBytes), duration).Err()
}

func (s *RedisDataStore) Remove(context stdcontext.Context, key string) error {
	return s.client.Del(context, key).Err()
}

func jsonToCoreValue(v interface{}) core.Value {
	switch val := v.(type) {
	case string:
		return core.ValText(val)
	case float64:
		return core.ValF64(val)
	case bool:
		return core.ValBool(val)
	case []interface{}:
		return core.ValNull() // simplified
	case map[string]interface{}:
		return core.ValNull() // simplified
	default:
		return core.ValNull()
	}
}
