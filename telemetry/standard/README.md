# Standard OpenTelemetry module for Go

This optional module creates a TeaQL-owned OpenTelemetry SDK at runtime startup.
Enabling it records supported TeaQL operations, masked SQL and business TraceChain
as traces, metrics and logs. Business methods need no SDK setup or manual spans.
The implementation is a local, unreleased candidate.

Import the module once in the application entry point:

```go
import _ "github.com/teaql/teaql-golang/telemetry/standard"
```

Enable export with trusted process configuration:

```bash
TEAQL_TELEMETRY_ENABLED=true
TEAQL_TELEMETRY_SERVICE_NAME=payment-service
TEAQL_TELEMETRY_OTLP_ENDPOINT=http://localhost:4318
```

When disabled, startup prints these three settings without creating exporters.
When enabled, startup prints validated effective settings and confirms full
recording and masking. Full recording does not mean raw values or lossless
export: bounded SDK queues can overflow and a collector can be unavailable.
`ExportHealth()` exposes failed spans, log records, metric batches and metric-data
objects, plus receiver rejection and partial/warning/invalid response counts.
`Queues` exposes offered, admitted, completed, outstanding, overflow, shutdown
and failed-handoff counts for spans/logs.

The module never replaces global SDK providers or shuts down a customer's SDK.
At application shutdown, call `context.RuntimeTelemetry().Shutdown(ctx)` once;
individual request contexts do not own the shared pipeline. Flush and shutdown
are bounded and fail-open. Business audit delivery remains independent.

## Local generated example

The example uses generated Q/E and audited mutation API bootstrap. It creates a
fresh SQLite database, verifies one seed and no repeated seed, then queries twice.
Ordinary SQL logs are disabled while standard telemetry remains enabled.

```bash
TEAQL_TELEMETRY_GO_BINARY=/path/to/go-1.25-or-newer \
  bash examples/trace-chain/verify-standard-telemetry.sh
```

The root runtime remains compatible with Go 1.18. This module needs Go 1.25 or
newer because it uses the modern official SDK, including its logs API. The
genproto compatibility floor in `go.mod` must remain present: removing it can
reintroduce package ambiguity between the runtime's legacy adapter dependencies
and the modern SDK when building consumers. All local examples resolve this
repository's runtime rather than a published package.

## Complete Order and graph examples

`examples/trace-chain/standard-order` uses generated Q/E for optional count,
forward details, reverse previews and facets. One invocation owns one root
through typed hydration. `examples/trace-chain/standard-graph` uses generated
audited Save for one order and two items. It checks normal commit, a real SQLite
rollback after an injected commit failure, and successful recovery. Explicit
child comments are retained; blank children inherit the parent reason. A root
remains active through committed audit and cleanup, without installing a span
or mutation ledger on shared UserContext.

The complete local gate runs both examples twice against retained databases,
verifies audited seed/repeated-seed behavior, and checks generated-library hashes:

```bash
TEAQL_TELEMETRY_GO_BINARY=/path/to/go-1.25-or-newer/bin/go \
  bash scripts/verify-examples.sh
```

## Typed scalar stream example

`examples/trace-chain/standard-stream` exercises complete consumption, intentional
early stop with `runtime.ErrStopStream`, context cancellation, callback error and
callback panic. Its runtime bridge owns one logical root through typed hydration,
actual delivery and cursor cleanup. Root spans, physical spans and SQL logs
record the same delivery count, excluding unconsumed rows from prefetched chunks.
An independent query inside a callback retains its own root. Original failures
remain unchanged; no application-level spans are required.

Count, facets and relationship enhancement cannot be silently discarded: typed
stream execution rejects them before SQL. Use the list API for those shapes.
Detailed cancellation/count completion is optional, preserving the original
Success/Failure adapter SPI. The complete local example gate includes this
stream fixture and runs twice without database cleanup.

## Export acknowledgements

Official SDK exporters keep record serialization. A TeaQL-owned HTTP client has
a three-second timeout, does not follow redirects and disables exporter retries.
Successful bodies are bounded to four MiB and decoded with official protobuf
bindings. Deprecated group wire types are rejected before decoding, including
unknown fields that bypass the decoder's message-recursion setting. Remote
response text, headers and transport errors do not enter SDK logging.

Partial success counts actual rejected spans/log records/metric points, not the
entire submitted batch. Zero-rejection warnings are visible but are not loss.
Negative or above-submitted rejection counts are invalid acknowledgements. Loss,
rejection and response counters are exported with fixed signal/reason labels;
failed metric-data objects and receiver-rejected metric points remain distinct.
An invalid acknowledgement cannot prove what the receiver stored. The generated
graph example prints a safe `GO_STANDARD_EXPORT_HEALTH` snapshot for local tests
and operations diagnostics without configuring spans in business methods.

## Queue capacity and shutdown

Each signal admits at most 2,048 records across queued, staged and currently
exporting batches, with batches of at most 128. SDK handoff and close share a
short monitor; network export does not hold it. Export completion releases
capacity. Overflow, submission after shutdown and failed handoff have separate
fixed loss counters. Directly ending the returned SDK span still goes through
admission exactly once.

Flush waits for accepted records and uses the official batch timer; it does not
enqueue SDK force-flush marker spans into the record budget. The generated graph
example supports `--queue-pressure --post-shutdown` for a controlled test with
1,500 actual Q/E reads, receiver release, recovery and business continuity.
Late losses after providers close remain available through `ExportHealth()`;
no extra metric snapshot is invented. Ordinary logging and audit stay independent.

Only successful SDK handoffs enter the flush prefix; a failed handoff cannot
satisfy an earlier record still waiting for export. Shutdown closes admission,
drains for at most five seconds, then uses the remainder of its six-second budget
for owned HTTP cancellation, attempt settlement and SDK cleanup. Queued records
are retired once; actual in-flight attempts have separate success/failure counts.
`Retired` and `Exporting` make these states visible in `ExportHealth().Queues`.
Retired batches are not sent after the cutoff. Final metrics follow span/log
settlement; unavailable receivers leave local counters, not an invented delivery.

The generated graph example also supports
`--queue-pressure --post-shutdown --shutdown-backlog` for a never-acknowledging
receiver. Real Q/E/audit stays correct, including a read after finite cleanup.
Actual failing-provider readback, enhanced streaming, advanced mixed/concurrent
graph and internal-artifact acceptance remain separate gates.
No public optional-module release is claimed.
