package nacos

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/cloud/core"
)

func TestNacosCloud(t *testing.T) {
	n := NewNacosCloud(&NacosConfig{ServerAddrs: []string{"localhost:8848"}})
	
	ctx := context.Background()
	instance := &core.ServiceInstance{ServiceId: "test-service"}
	
	err := n.Register(ctx, instance)
	assert.NoError(t, err)
	
	err = n.Deregister(ctx, instance)
	assert.NoError(t, err)
	
	instances, err := n.GetInstances(ctx, "test-service")
	assert.Nil(t, instances)
	assert.NoError(t, err)
	
	cfg, err := n.GetConfig(ctx, "data", "group")
	assert.Equal(t, "", cfg)
	assert.NoError(t, err)
	
	assert.Equal(t, core.Up, n.Health())
}
