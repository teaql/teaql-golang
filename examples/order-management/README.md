# Order Management — Go + SQLite

Run from a terminal; no database server, fixture DB, model input, or generator installation is needed:

```bash
cd examples/order-management/golang-app-console
go run .
```

The first run creates `../.local/order.db`, ensures schema from generated metadata, seeds through generated entity APIs, performs a governed query, and writes an audited preset. The second run demonstrates idempotency.

Read `golang-app-console/main.go` first (handwritten), then `golang-lib-core/q.go`, `customer_order/request.go`, and `customer_order/entity.go` (generated). The library is a standard Go module; the console is a separate module with a local `replace` only for this source-tree example. TeaQL itself resolves from its published GitHub module version.

## Verify the first result

The trace must contain both comment and purpose, followed by `WEB-2026-001`, `2026-08-12`, and `129.95`. The first run emits safe application audit events; the second adds no duplicates.

## Customize it

Change `WithOrderNumberContaining`, add generated ordering, or select generated relations in `main.go`. Read the generated request before using another operator. Handwritten policy stays in `golang-app-console`; regenerate `golang-lib-core`. The shared model is not required to run the example.
### Materialized-list hard limit

`ExecuteForList` protects the service by applying a default hard limit of 10,000 rows. A requested page size above that ceiling fails explicitly. Trusted application code can call `HardLimit(...)` to override the outer-query ceiling. **Caution:** most applications should not override it; do so only for a reviewed, exceptional requirement. This setting does not describe streaming execution.

### Continuous browsing optimization

For an explicitly browse-only screen ordered solely by the unique `id`, local server code can
opt in to transparent seek pagination without changing `ExecuteForList`:

```go
orders, err := Q.CustomerOrders().
    OrderByIdDesc().
    Offset(1000).
    Limit(20).
    OptimizeForContinuousPageFetchWith("recent-orders", 60).
    Comment("Load the next browse page").
    Purpose("Browse recent orders").
    ExecuteForList(ctx)
```

After a preceding page has registered its boundary, the runtime can replace the next offset with
`id < boundary` for descending order, or `id > boundary` for ascending order. A missing, expired,
invalid, or unavailable cursor safely falls back to the original offset query. Inspect
`ctx.ContinuousPagePlan()` and `ctx.ContinuousPageCursorID()` when debugging.

**Caution:** this is a hallucination-tolerant performance optimization for continuous browsing.
Concurrent inserts or updates can make page membership differ from exact offset pagination. Do not
use it for business workflows, reconciliation, exports, or other exact-result processing. It is a
local runtime hint and cannot be supplied or overridden through TeaQL federation JSON.

### Streaming large root queries

`ExecuteForStream` invokes the callback once per generated entity while the provider cursor remains open:

```go
err := request.
	Comment("export orders").
	Purpose("reviewed export").
	ExecuteForStream(
		ctx,
		500,
		func(order *customerorder.CustomerOrder) error {
			return writeOrder(order)
		},
	)
```

Returning an error stops iteration and closes the cursor. The chunk size is an internal fetch bound. **Caution:** normally keep the default-scale value (1,000). Relation or aggregate enhancement is rejected for streaming; use a root query or `ExecuteForList`. Ordinary TFP federation does not emulate a stream.
