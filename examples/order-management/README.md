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
