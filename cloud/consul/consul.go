package consul

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/teaql/teaql-golang/cloud/core"
)

type ConsulConfig struct {
	ServerAddr string
	Token      *string
	HTTPClient *http.Client
}

type ConsulCloud struct {
	config *ConsulConfig
	client *http.Client
}

func NewConsulCloud(config *ConsulConfig) *ConsulCloud {
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &ConsulCloud{config: config, client: client}
}

func (c *ConsulCloud) endpoint(path string) string {
	base := c.config.ServerAddr
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}
	return strings.TrimRight(base, "/") + path
}

func (c *ConsulCloud) request(context stdcontext.Context, method, path string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(context, method, c.endpoint(path), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.config.Token != nil {
		req.Header.Set("X-Consul-Token", *c.config.Token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	payload, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("consul %s %s failed: %s: %s", method, path, resp.Status, payload)
	}
	return payload, nil
}

func (c *ConsulCloud) Register(context stdcontext.Context, instance *core.ServiceInstance) error {
	_, err := c.request(context, http.MethodPut, "/v1/agent/service/register", map[string]any{
		"ID": instance.ServiceId, "Name": instance.ServiceId, "Address": instance.Host,
		"Port": instance.Port, "Meta": instance.Metadata,
	})
	return err
}

func (c *ConsulCloud) Deregister(context stdcontext.Context, instance *core.ServiceInstance) error {
	_, err := c.request(context, http.MethodPut, "/v1/agent/service/deregister/"+url.PathEscape(instance.ServiceId), nil)
	return err
}

func (c *ConsulCloud) GetInstances(context stdcontext.Context, serviceId string) ([]*core.ServiceInstance, error) {
	payload, err := c.request(context, http.MethodGet, "/v1/health/service/"+url.PathEscape(serviceId)+"?passing=true", nil)
	if err != nil {
		return nil, err
	}
	var result []struct {
		Service struct {
			ID      string            `json:"ID"`
			Address string            `json:"Address"`
			Port    int               `json:"Port"`
			Meta    map[string]string `json:"Meta"`
		} `json:"Service"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, fmt.Errorf("decode consul instances: %w", err)
	}
	instances := make([]*core.ServiceInstance, 0, len(result))
	for _, item := range result {
		instances = append(instances, &core.ServiceInstance{ServiceId: item.Service.ID, Host: item.Service.Address, Port: item.Service.Port, Metadata: item.Service.Meta})
	}
	return instances, nil
}

func (c *ConsulCloud) Health() core.HealthStatus {
	timeout := c.client.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	context, cancel := stdcontext.WithTimeout(stdcontext.Background(), timeout)
	defer cancel()
	payload, err := c.request(context, http.MethodGet, "/v1/status/leader", nil)
	if err != nil || strings.TrimSpace(string(payload)) == `""` {
		return core.Down
	}
	return core.Up
}
