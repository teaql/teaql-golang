package starter

import (
	"github.com/teaql/teaql-golang/cloud/core"
	"github.com/teaql/teaql-golang/cloud/nacos"
	"github.com/teaql/teaql-golang/cloud/consul"
)

type CloudStarter struct {
	Registry  core.ServiceRegistry
	Discovery core.ServiceDiscovery
	Config    core.ConfigSource
}

func InitNacos(config *nacos.NacosConfig) *CloudStarter {
	nc := nacos.NewNacosCloud(config)
	return &CloudStarter{
		Registry:  nc,
		Discovery: nc,
		Config:    nc,
	}
}

func InitConsul(config *consul.ConsulConfig) *CloudStarter {
	cc := consul.NewConsulCloud(config)
	return &CloudStarter{
		Registry:  cc,
		// Discovery & Config omitted for briefness
	}
}
