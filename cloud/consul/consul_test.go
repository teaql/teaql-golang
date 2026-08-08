package consul

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/cloud/core"
)

func TestConsulCloud(t *testing.T) {
	c := NewConsulCloud(&ConsulConfig{ServerAddr: "localhost:8500"})
	
	ctx := context.Background()
	instance := &core.ServiceInstance{ServiceId: "test-service"}
	
	err := c.Register(ctx, instance)
	assert.NoError(t, err)
	
	err = c.Deregister(ctx, instance)
	assert.NoError(t, err)
	
	assert.Equal(t, core.Up, c.Health())
}
