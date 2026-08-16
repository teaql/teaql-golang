package consul

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teaql/teaql-golang/cloud/core"
)

func TestConsulCloudUsesRealHTTPAPI(t *testing.T) {
	requests := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.Method+" "+r.URL.Path]++
		switch r.URL.Path {
		case "/v1/agent/service/register":
			var body map[string]any
			if json.NewDecoder(r.Body).Decode(&body) != nil || body["ID"] != "orders" {
				http.Error(w, "bad registration", http.StatusBadRequest)
			}
		case "/v1/agent/service/deregister/orders":
		case "/v1/health/service/orders":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"Service":{"ID":"orders-1","Address":"10.0.0.8","Port":8081,"Meta":{"zone":"b"}}}]`))
		case "/v1/status/leader":
			_, _ = w.Write([]byte(`"127.0.0.1:8300"`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	token := "secret"
	cloud := NewConsulCloud(&ConsulConfig{ServerAddr: server.URL, Token: &token, HTTPClient: server.Client()})
	instance := &core.ServiceInstance{ServiceId: "orders", Host: "10.0.0.8", Port: 8081}
	ctx := context.Background()
	if err := cloud.Register(ctx, instance); err != nil {
		t.Fatal(err)
	}
	if err := cloud.Deregister(ctx, instance); err != nil {
		t.Fatal(err)
	}
	instances, err := cloud.GetInstances(ctx, "orders")
	if err != nil || len(instances) != 1 || instances[0].ServiceId != "orders-1" {
		t.Fatalf("instances=%v err=%v", instances, err)
	}
	if cloud.Health() != core.Up {
		t.Fatal("expected UP")
	}
	if requests["PUT /v1/agent/service/register"] != 1 {
		t.Fatalf("requests=%v", requests)
	}
}
