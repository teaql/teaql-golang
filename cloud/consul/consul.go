package consul

import (
	"context"
	"fmt"
	"github.com/teaql/teaql-golang/cloud/core"
)

type ConsulConfig struct {
	ServerAddr string
	Token      *string
}

type ConsulCloud struct {
	config *ConsulConfig
}

func NewConsulCloud(config *ConsulConfig) *ConsulCloud {
	return &ConsulCloud{
		config: config,
	}
}

func (c *ConsulCloud) Register(ctx context.Context, instance *core.ServiceInstance) error {
	fmt.Printf("Registering instance %s to Consul at %v\n", instance.ServiceId, c.config.ServerAddr)
	return nil
}

func (c *ConsulCloud) Deregister(ctx context.Context, instance *core.ServiceInstance) error {
	return nil
}

func (c *ConsulCloud) Health() core.HealthStatus {
	return core.Up
}
