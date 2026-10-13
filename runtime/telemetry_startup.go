package runtime

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const TelemetryEnabledEnvironment = "TEAQL_TELEMETRY_ENABLED"
const TelemetryServiceNameEnvironment = "TEAQL_TELEMETRY_SERVICE_NAME"
const TelemetryEndpointEnvironment = "TEAQL_TELEMETRY_OTLP_ENDPOINT"
const telemetryHints = "TEAQL_TELEMETRY_ENABLED=true TEAQL_TELEMETRY_SERVICE_NAME=<application-name> TEAQL_TELEMETRY_OTLP_ENDPOINT=<collector-url>"

type TelemetryStartupConfiguration struct {
	Enabled                          bool
	ServiceName, Endpoint, ErrorCode string
}

var telemetryServicePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,120}$`)

func ReadTelemetryStartupConfiguration(environment map[string]string) TelemetryStartupConfiguration {
	flag := strings.TrimSpace(environment[TelemetryEnabledEnvironment])
	if flag == "" || strings.EqualFold(flag, "false") {
		return TelemetryStartupConfiguration{}
	}
	if !strings.EqualFold(flag, "true") {
		return TelemetryStartupConfiguration{ErrorCode: "invalid-enable"}
	}
	service := strings.TrimSpace(environment[TelemetryServiceNameEnvironment])
	if !telemetryServicePattern.MatchString(service) {
		return TelemetryStartupConfiguration{ErrorCode: "invalid-service"}
	}
	endpoint, err := url.Parse(environment[TelemetryEndpointEnvironment])
	if err != nil || endpoint == nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.ForceQuery || strings.Contains(environment[TelemetryEndpointEnvironment], "#") {
		return TelemetryStartupConfiguration{ErrorCode: "invalid-endpoint"}
	}
	return TelemetryStartupConfiguration{Enabled: true, ServiceName: service, Endpoint: strings.TrimRight(endpoint.String(), "/")}
}
func (config TelemetryStartupConfiguration) Message() string {
	if config.ErrorCode != "" {
		return "TeaQL telemetry: disabled; configuration=" + config.ErrorCode + ". " + telemetryHints
	}
	if !config.Enabled {
		return "TeaQL telemetry: supported, disabled. Enable full masked diagnostics with " + telemetryHints
	}
	return fmt.Sprintf("TeaQL telemetry: starting standard module; full recording, masked SQL; %s=true %s=%s %s=%s", TelemetryEnabledEnvironment, TelemetryServiceNameEnvironment, config.ServiceName, TelemetryEndpointEnvironment, config.Endpoint)
}
func TelemetryStartupNotice(message string) { _, _ = fmt.Fprintln(os.Stderr, message) }

// Optional modules register a factory; registration creates no SDK resources.
type StandardTelemetryFactory func(context.Context, TelemetryStartupConfiguration) (RuntimeTelemetry, error)

var standardTelemetryState = struct {
	sync.Mutex
	factory   StandardTelemetryFactory
	pipelines map[string]RuntimeTelemetry
	announced map[string]bool
}{pipelines: map[string]RuntimeTelemetry{}, announced: map[string]bool{}}

func RegisterStandardTelemetryFactory(factory StandardTelemetryFactory) {
	standardTelemetryState.Lock()
	defer standardTelemetryState.Unlock()
	standardTelemetryState.factory = factory
}
func initializeStandardTelemetry(environment map[string]string) RuntimeTelemetry {
	config := ReadTelemetryStartupConfiguration(environment)
	standardTelemetryState.Lock()
	defer standardTelemetryState.Unlock()
	message := config.Message()
	if !standardTelemetryState.announced[message] {
		standardTelemetryState.announced[message] = true
		TelemetryStartupNotice(message)
	}
	if !config.Enabled {
		return NoopRuntimeTelemetry{}
	}
	if standardTelemetryState.factory == nil {
		missing := "TeaQL telemetry: standard module unavailable; import github.com/teaql/teaql-golang/telemetry/standard. " + telemetryHints
		if !standardTelemetryState.announced[missing] {
			standardTelemetryState.announced[missing] = true
			TelemetryStartupNotice(missing)
		}
		return NoopRuntimeTelemetry{}
	}
	key := config.ServiceName + "\x00" + config.Endpoint
	if pipeline := standardTelemetryState.pipelines[key]; pipeline != nil {
		closed, ok := pipeline.(interface{ IsShutdown() bool })
		if !ok || !closed.IsShutdown() {
			return pipeline
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type result struct {
		pipeline RuntimeTelemetry
		err      error
	}
	completed := make(chan result, 1)
	factory := standardTelemetryState.factory
	go func() {
		outcome := result{}
		defer func() {
			if recover() != nil {
				outcome.err = fmt.Errorf("initialization-failed")
			}
			if ctx.Err() != nil && outcome.pipeline != nil {
				cleanup, stop := context.WithTimeout(context.Background(), 6*time.Second)
				defer stop()
				_ = outcome.pipeline.Shutdown(cleanup)
			}
			completed <- outcome
		}()
		outcome.pipeline, outcome.err = factory(ctx, config)
	}()
	select {
	case outcome := <-completed:
		if outcome.err == nil && outcome.pipeline != nil {
			standardTelemetryState.pipelines[key] = outcome.pipeline
			TelemetryStartupNotice("TeaQL telemetry: enabled; full recording, traces, metrics, logs and masked SQL.")
			return outcome.pipeline
		}
	case <-ctx.Done():
	}
	TelemetryStartupNotice("TeaQL telemetry: standard module initialization failed; business continues without export.")
	return NoopRuntimeTelemetry{}
}
func defaultRuntimeTelemetry() RuntimeTelemetry {
	environment := map[string]string{}
	for _, name := range []string{TelemetryEnabledEnvironment, TelemetryServiceNameEnvironment, TelemetryEndpointEnvironment} {
		environment[name] = os.Getenv(name)
	}
	return initializeStandardTelemetry(environment)
}
