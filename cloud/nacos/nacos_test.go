package nacos

import (
	stdcontext "context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teaql/teaql-golang/cloud/core"
)

func TestNacosCloudUsesRealHTTPAPI(t *testing.T) {
	requests := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.Method+" "+r.URL.Path]++
		switch r.URL.Path {
		case "/nacos/v1/ns/instance":
			if err := r.ParseForm(); err != nil || r.Form.Get("serviceName") != "orders" || r.Form.Get("ip") != "10.0.0.7" {
				http.Error(w, "bad instance request", http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, "ok")
		case "/nacos/v1/ns/instance/list":
			fmt.Fprint(w, `{"hosts":[{"ip":"10.0.0.7","port":8080,"healthy":true,"metadata":{"zone":"a"}}]}`)
		case "/nacos/v1/cs/configs":
			fmt.Fprint(w, "feature=true")
		case "/nacos/v1/console/health/readiness":
			fmt.Fprint(w, `{"status":"UP"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cloud := NewNacosCloud(&NacosConfig{ServerAddrs: []string{server.URL}, NamespaceId: "tenant-a", Group: "APP", HTTPClient: server.Client()})
	instance := &core.ServiceInstance{ServiceId: "orders", Host: "10.0.0.7", Port: 8080, Metadata: map[string]string{"zone": "a"}}
	context := stdcontext.Background()
	if err := cloud.Register(context, instance); err != nil {
		t.Fatal(err)
	}
	if err := cloud.Deregister(context, instance); err != nil {
		t.Fatal(err)
	}
	instances, err := cloud.GetInstances(context, "orders")
	if err != nil || len(instances) != 1 || instances[0].Host != "10.0.0.7" {
		t.Fatalf("instances=%v err=%v", instances, err)
	}
	config, err := cloud.GetConfig(context, "orders.yaml", "APP")
	if err != nil || config != "feature=true" {
		t.Fatalf("config=%q err=%v", config, err)
	}
	if cloud.Health() != core.Up {
		t.Fatal("expected UP")
	}
	if requests["POST /nacos/v1/ns/instance"] != 1 || requests["DELETE /nacos/v1/ns/instance"] != 1 {
		t.Fatalf("requests=%v", requests)
	}
}
