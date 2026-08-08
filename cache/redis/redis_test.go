package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
)

func TestRedisDataStore(t *testing.T) {
	s := miniredis.RunT(t)

	ds, err := NewRedisDataStore("redis://" + s.Addr())
	assert.NoError(t, err)
	assert.NotNil(t, ds.Client())

	ctx := context.Background()

	// Test Get empty
	val, ok := ds.Get(ctx, "nonexistent")
	assert.False(t, ok)
	assert.Equal(t, core.ValNull(), val)

	// Test Put
	err = ds.Put(ctx, "key1", core.ValText("value1"), nil)
	assert.NoError(t, err)

	// Test Get existing
	val, ok = ds.Get(ctx, "key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val.V)

	// Test Remove
	err = ds.Remove(ctx, "key1")
	assert.NoError(t, err)

	// Test Get after remove
	val, ok = ds.Get(ctx, "key1")
	assert.False(t, ok)

	// Test timeout
	timeout := uint64(10)
	err = ds.Put(ctx, "key2", core.ValF64(123), &timeout)
	assert.NoError(t, err)
	
	// Test Invalid json format
	s.Set("invalid_json", "{")
	val, ok = ds.Get(ctx, "invalid_json")
	assert.False(t, ok)
	
	// Test parse failure for redis URL
	_, err = NewRedisDataStore("::invalid_url")
	assert.Error(t, err)
}

func TestJsonToCoreValue(t *testing.T) {
	assert.Equal(t, core.ValText("test"), jsonToCoreValue("test"))
	assert.Equal(t, core.ValF64(12.3), jsonToCoreValue(12.3))
	assert.Equal(t, core.ValBool(true), jsonToCoreValue(true))
	assert.Equal(t, core.ValNull(), jsonToCoreValue([]interface{}{}))
	assert.Equal(t, core.ValNull(), jsonToCoreValue(map[string]interface{}{}))
	assert.Equal(t, core.ValNull(), jsonToCoreValue(nil))
}

func TestRedisDataStore_MarshalError(t *testing.T) {
	s := miniredis.RunT(t)
	ds, _ := NewRedisDataStore("redis://" + s.Addr())
	err := ds.Put(context.Background(), "key", core.Value{V: make(chan int)}, nil)
	assert.Error(t, err)
}

func TestRedisDataStore_PutError(t *testing.T) {
	ds, _ := NewRedisDataStore("redis://localhost:9999") // invalid server
	err := ds.Put(context.Background(), "key", core.ValText("a"), nil)
	assert.Error(t, err) // connection refused
}
