package nacos

import (
	"context"
	"fmt"
	"github.com/teaql/teaql-golang/cloud/core"
)

type NacosConfig struct {
	ServerAddrs []string
	NamespaceId string
	Group       string
}

type NacosCloud struct {
	config *NacosConfig
}

func NewNacosCloud(config *NacosConfig) *NacosCloud {
	return &NacosCloud{
		config: config,
	}
}

func (n *NacosCloud) Register(ctx context.Context, instance *core.ServiceInstance) error {
	// Dummy implementation for now
	fmt.Printf("Registering instance %s to Nacos at %v\n", instance.ServiceId, n.config.ServerAddrs)
	return nil
}

func (n *NacosCloud) Deregister(ctx context.Context, instance *core.ServiceInstance) error {
	return nil
}

func (n *NacosCloud) GetInstances(ctx context.Context, serviceId string) ([]*core.ServiceInstance, error) {
	return nil, nil
}

func (n *NacosCloud) GetConfig(ctx context.Context, dataId, group string) (string, error) {
	return "", nil
}

func (n *NacosCloud) Health() core.HealthStatus {
	return core.Up
}
