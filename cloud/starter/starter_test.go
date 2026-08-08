package starter

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/cloud/consul"
	"github.com/teaql/teaql-golang/cloud/nacos"
)

func TestInitNacos(t *testing.T) {
	cs := InitNacos(&nacos.NacosConfig{ServerAddrs: []string{"localhost:8848"}})
	assert.NotNil(t, cs)
	assert.NotNil(t, cs.Registry)
	assert.NotNil(t, cs.Discovery)
	assert.NotNil(t, cs.Config)
}

func TestInitConsul(t *testing.T) {
	cs := InitConsul(&consul.ConsulConfig{ServerAddr: "localhost:8500"})
	assert.NotNil(t, cs)
	assert.NotNil(t, cs.Registry)
	assert.Nil(t, cs.Discovery)
	assert.Nil(t, cs.Config)
}
