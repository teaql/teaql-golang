package nacos

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/teaql/teaql-golang/cloud/core"
)

type NacosConfig struct {
	ServerAddrs []string
	NamespaceId string
	Group       string
	Username    string
	Password    string
	HTTPClient  *http.Client
}

type NacosCloud struct {
	config *NacosConfig
	client *http.Client
}

func NewNacosCloud(config *NacosConfig) *NacosCloud {
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &NacosCloud{config: config, client: client}
}

func (n *NacosCloud) endpoint(path string) (string, error) {
	if len(n.config.ServerAddrs) == 0 {
		return "", fmt.Errorf("nacos server address is required")
	}
	base := n.config.ServerAddrs[0]
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}
	return strings.TrimRight(base, "/") + path, nil
}

func (n *NacosCloud) serviceValues(instance *core.ServiceInstance) url.Values {
	values := url.Values{
		"serviceName": {instance.ServiceId},
		"ip":          {instance.Host},
		"port":        {strconv.Itoa(instance.Port)},
		"groupName":   {n.group()},
		"namespaceId": {n.config.NamespaceId},
		"ephemeral":   {"true"},
	}
	if instance.Metadata != nil {
		if encoded, err := json.Marshal(instance.Metadata); err == nil {
			values.Set("metadata", string(encoded))
		}
	}
	return values
}

func (n *NacosCloud) group() string {
	if n.config.Group == "" {
		return "DEFAULT_GROUP"
	}
	return n.config.Group
}

func (n *NacosCloud) request(context stdcontext.Context, method, path string, values url.Values) ([]byte, error) {
	endpoint, err := n.endpoint(path)
	if err != nil {
		return nil, err
	}
	var body io.Reader
	if method == http.MethodGet || method == http.MethodDelete {
		endpoint += "?" + values.Encode()
	} else {
		body = strings.NewReader(values.Encode())
	}
	req, err := http.NewRequestWithContext(context, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if n.config.Username != "" {
		req.SetBasicAuth(n.config.Username, n.config.Password)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	payload, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("nacos %s %s failed: %s: %s", method, path, resp.Status, payload)
	}
	return payload, nil
}

func (n *NacosCloud) Register(context stdcontext.Context, instance *core.ServiceInstance) error {
	_, err := n.request(context, http.MethodPost, "/nacos/v1/ns/instance", n.serviceValues(instance))
	return err
}

func (n *NacosCloud) Deregister(context stdcontext.Context, instance *core.ServiceInstance) error {
	_, err := n.request(context, http.MethodDelete, "/nacos/v1/ns/instance", n.serviceValues(instance))
	return err
}

func (n *NacosCloud) GetInstances(context stdcontext.Context, serviceId string) ([]*core.ServiceInstance, error) {
	payload, err := n.request(context, http.MethodGet, "/nacos/v1/ns/instance/list", url.Values{
		"serviceName": {serviceId}, "groupName": {n.group()}, "namespaceId": {n.config.NamespaceId}, "healthyOnly": {"true"},
	})
	if err != nil {
		return nil, err
	}
	var result struct {
		Hosts []struct {
			IP       string            `json:"ip"`
			Port     int               `json:"port"`
			Healthy  bool              `json:"healthy"`
			Metadata map[string]string `json:"metadata"`
		} `json:"hosts"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, fmt.Errorf("decode nacos instances: %w", err)
	}
	instances := make([]*core.ServiceInstance, 0, len(result.Hosts))
	for _, host := range result.Hosts {
		if host.Healthy {
			instances = append(instances, &core.ServiceInstance{ServiceId: serviceId, Host: host.IP, Port: host.Port, Metadata: host.Metadata})
		}
	}
	return instances, nil
}

func (n *NacosCloud) GetConfig(context stdcontext.Context, dataId, group string) (string, error) {
	if group == "" {
		group = n.group()
	}
	payload, err := n.request(context, http.MethodGet, "/nacos/v1/cs/configs", url.Values{
		"dataId": {dataId}, "group": {group}, "tenant": {n.config.NamespaceId},
	})
	return string(payload), err
}

func (n *NacosCloud) Health() core.HealthStatus {
	timeout := n.client.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	context, cancel := stdcontext.WithTimeout(stdcontext.Background(), timeout)
	defer cancel()
	_, err := n.request(context, http.MethodGet, "/nacos/v1/console/health/readiness", url.Values{})
	if err != nil {
		return core.Down
	}
	return core.Up
}
