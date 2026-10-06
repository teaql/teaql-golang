<!-- ephemeral -->

# TeaQL Go Runtime Customization

Generated `ServiceRuntimeFromEnv` owns connection and schema startup. The application composition
root adds policy, trusted tenant and App audit sink; request/federation JSON cannot set them.

```go
package app

import (
	"fmt"
	"strings"

	"github.com/teaql/teaql-golang/runtime"
)

func ConfiguredRuntime(
	base *runtime.UserContext,
	requestPolicy runtime.RequestPolicy,
	trustedTenant string,
	appAuditSink runtime.AppAuditEventSink,
) (*runtime.UserContext, error) {
	if base == nil {
		return nil, fmt.Errorf("generated service runtime is required")
	}
	if strings.TrimSpace(trustedTenant) == "" {
		return nil, fmt.Errorf("trusted tenant is required")
	}
	base.WithRequestPolicy(requestPolicy).
		WithTrustedTenant(trustedTenant).
		WithAppAuditEventSink(appAuditSink)
	if err := base.RuntimeReadiness(); err != nil {
		return nil, err
	}
	return base, nil
}

var governanceKeys = map[string]struct{}{
	"tenant": {}, "trustedTenant": {}, "provider": {}, "dataService": {},
	"requestPolicy": {}, "auditSink": {}, "appAuditSink": {},
	"hardLimit": {}, "continuousPage": {},
}

func RejectGovernanceOverride(value any) error {
	switch value := value.(type) {
	case map[string]any:
		for key, nested := range value {
			if _, forbidden := governanceKeys[key]; forbidden {
				return fmt.Errorf("untrusted governance override: %s", key)
			}
			if err := RejectGovernanceOverride(nested); err != nil { return err }
		}
	case []any:
		for _, nested := range value {
			if err := RejectGovernanceOverride(nested); err != nil { return err }
		}
	}
	return nil
}
```

Pass the context returned by `ConfiguredRuntime` as the only context argument to generated queries
and audited mutations. `RuntimeReadiness` validates generated startup state and pings providers that
expose `PingContext`. Raw audit remains module-owned and App audit receives only safe events.

## Runtime telemetry

Observability is optional and application-owned. Construct
`telemetry/opentelemetry.RuntimeTelemetry` with `opentelemetry.New`, optionally
add the application log bridge with `WithLogEmitter`, then call
`context.WithRuntimeTelemetry(telemetry)`. Keep `runtime.NoopRuntimeTelemetry{}`
when absent. The application owns bounded processors, OTLP exporters, `Flush`
and `Shutdown`; exporter failure must never change business results. Telemetry
setup does not call schema readiness.
TeaQL derives `teaql.error.category` from the native error type. Sampling never
controls or replaces App Audit Sink delivery. Do not generate a Collector,
additional exporters, auto-discovery, or a telemetry configuration DSL.

---

## TeaQL seven-language assist contract

Apply the verified Rust semantic ceiling while using only the exact GOLANG generated and
runtime APIs. Discover APIs through the generated application AGENTS.md and progressive
model-aware Assist. Do not inspect generated domain-library source.

- Do not create plurals by appending `s` or `es`; use the centralized generated plural.
- Human and non-human entities use different generated predicate vocabularies. Preserve
  forms such as “who are active” and “whose email is”; never infer them from English.
- Configure filters, projection, paging, and other query options before `purpose(...)`.
  Comment may appear anywhere in the chain. Purpose enters the executable stage; execution
  requires both values, but comment does not have to immediately precede purpose.
- Every execute/list/stream and every save accepts exactly one context argument:
  `UserContext`. Name that argument `context`, never `runtime`; data services and global
  policy are injected when the context is built. Reserve `runtime` for process-level
  runtime ownership, provider/pool setup, and module assembly.
- Tenant, merchant, identity, permissions, request policy, purpose policy, hard limit,
  and continuous-page cursor policy come only from trusted context, never dynamic JSON or TFP.
- If the required operation is absent after current entity/action and required field
  Assist, stop that path and report MISSING_ASSIST. Do not guess an API or search the
  generated library as a fallback.
- Create each application-owned source file once. After its first compile attempt,
  repair only the smallest block identified by the exact compiler or test diagnostic.
  Preserve unrelated code; do not rewrite the complete file as an error-recovery loop.
- Before a repair that would replace more than 25% of an existing application file,
  stop and report LARGE_REWRITE_REQUEST with the file, exact diagnostic, reason, and
  estimated scope. Initial creation and model-driven regeneration are not repairs.

Capability: `runtime-custom`.

- Keep trusted dependencies and global runtime policy in UserContext initialization.
  Custom providers, policy hooks, and audit sinks must not add execute/save arguments.
- Preserve immutable row audit events and a separate customizable App Audit Sink.
  Include health, integration, and negative governance tests for every customization.
