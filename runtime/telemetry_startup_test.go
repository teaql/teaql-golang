package runtime

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

func telemetryTestEnvironment(endpoint string) map[string]string {
	return map[string]string{
		TelemetryEnabledEnvironment: "true", TelemetryServiceNameEnvironment: "orders", TelemetryEndpointEnvironment: endpoint,
	}
}
func TestTelemetryStartupDisabledGuidanceDoesNotEchoUnusedSettings(t *testing.T) {
	config := ReadTelemetryStartupConfiguration(map[string]string{TelemetryServiceNameEnvironment: "PRIVATE-SERVICE"})
	if config.Enabled || strings.Contains(config.Message(), "PRIVATE-SERVICE") {
		t.Fatal(config)
	}
	for _, name := range []string{TelemetryEnabledEnvironment, TelemetryServiceNameEnvironment, TelemetryEndpointEnvironment} {
		if !strings.Contains(config.Message(), name) {
			t.Fatal("missing hint", name)
		}
	}
}
func TestTelemetryStartupRejectsSecretURLsWithoutEcho(t *testing.T) {
	for _, endpoint := range []string{"http://SECRET:password@localhost:4318", "http://localhost:4318?token=SECRET", "http://localhost:4318?", "http://localhost:4318/#SECRET", "file:///SECRET", "http://@localhost:4318"} {
		config := ReadTelemetryStartupConfiguration(telemetryTestEnvironment(endpoint))
		if config.Enabled || config.ErrorCode != "invalid-endpoint" || strings.Contains(config.Message(), "SECRET") {
			t.Fatal(config.Message())
		}
	}
}
func TestTelemetryStartupValidatesSwitchServiceAndNormalizesEndpoint(t *testing.T) {
	config := ReadTelemetryStartupConfiguration(telemetryTestEnvironment("http://localhost:4318/"))
	if !config.Enabled || config.Endpoint != "http://localhost:4318" || !strings.Contains(config.Message(), "full recording") {
		t.Fatal(config)
	}
	input := telemetryTestEnvironment("http://localhost:4318")
	input[TelemetryEnabledEnvironment] = "SECRET"
	if ReadTelemetryStartupConfiguration(input).ErrorCode != "invalid-enable" {
		t.Fatal("switch must be explicit")
	}
	input[TelemetryEnabledEnvironment] = "true"
	input[TelemetryServiceNameEnvironment] = "SECRET / service"
	if ReadTelemetryStartupConfiguration(input).ErrorCode != "invalid-service" {
		t.Fatal("invalid service accepted")
	}
}

type startupTestTelemetry struct {
	NoopRuntimeTelemetry
	closed int32
}

func (p *startupTestTelemetry) IsShutdown() bool { return atomic.LoadInt32(&p.closed) != 0 }
func (p *startupTestTelemetry) Shutdown(context.Context) error {
	atomic.StoreInt32(&p.closed, 1)
	return nil
}
func TestTelemetryStartupDefersFactoryAndSharesPipelineAcrossContexts(t *testing.T) {
	standardTelemetryState.Lock()
	previous := standardTelemetryState.factory
	standardTelemetryState.Unlock()
	t.Cleanup(func() { RegisterStandardTelemetryFactory(previous) })
	var calls int32
	RegisterStandardTelemetryFactory(func(context.Context, TelemetryStartupConfiguration) (RuntimeTelemetry, error) {
		atomic.AddInt32(&calls, 1)
		return &startupTestTelemetry{}, nil
	})
	initializeStandardTelemetry(map[string]string{})
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatal("disabled startup created SDK resources")
	}
	input := telemetryTestEnvironment("http://localhost:4318")
	input[TelemetryServiceNameEnvironment] = t.Name()
	first := initializeStandardTelemetry(input)
	second := initializeStandardTelemetry(input)
	if first != second || atomic.LoadInt32(&calls) != 1 {
		t.Fatal("SDK must not be constructed per context")
	}
	_ = first.Shutdown(context.Background())
	third := initializeStandardTelemetry(input)
	if third == first || atomic.LoadInt32(&calls) != 2 {
		t.Fatal("a closed pipeline must not be reused")
	}
}
func TestTelemetryStartupFactoryFailureIsFailOpen(t *testing.T) {
	standardTelemetryState.Lock()
	previous := standardTelemetryState.factory
	standardTelemetryState.Unlock()
	t.Cleanup(func() { RegisterStandardTelemetryFactory(previous) })
	RegisterStandardTelemetryFactory(func(context.Context, TelemetryStartupConfiguration) (RuntimeTelemetry, error) {
		return nil, errors.New("SECRET-FAILURE")
	})
	input := telemetryTestEnvironment("http://localhost:4318")
	input[TelemetryServiceNameEnvironment] = t.Name()
	if _, ok := initializeStandardTelemetry(input).(NoopRuntimeTelemetry); !ok {
		t.Fatal("failed initialization must preserve business continuity")
	}
}
