package core

import "context"

type ServiceInstance struct {
	ServiceId string
	Host      string
	Port      int
	Secure    bool
	Metadata  map[string]string
}

type ServiceRegistry interface {
	Register(context context.Context, instance *ServiceInstance) error
	Deregister(context context.Context, instance *ServiceInstance) error
}

type ServiceDiscovery interface {
	GetInstances(context context.Context, serviceId string) ([]*ServiceInstance, error)
}

type ConfigSource interface {
	GetConfig(context context.Context, dataId, group string) (string, error)
}

type HealthStatus string

const (
	Up           HealthStatus = "UP"
	Down         HealthStatus = "DOWN"
	OutOfService HealthStatus = "OUT_OF_SERVICE"
	Unknown      HealthStatus = "UNKNOWN"
)

type HealthIndicator interface {
	Health() HealthStatus
}
