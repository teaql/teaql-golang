<!-- ephemeral -->

Please help me complete the query (Query) service business code for the `Order Item` object.

### Standard Query Example (Reference)
```go
package assist_query

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	. "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/order_item"
)

func QueryExample(context *runtime.UserContext, id uint64) (*core.SmartList[*order_item.OrderItem], error) {
	listResult, err := Q.OrderItemsMinimal().
		WithIdIs(id).

		OrderByIdAsc().
		Limit(20).
		Comment("what: retrieve Order Item by generated ID filter").
		Purpose("why: answer a bounded business query").
		ExecuteForList(context)

	if err != nil {
		return nil, err
	}

	return listResult, nil
}
```

### Your Task
Please implement the real query logic for `Order Item` based on my specific business needs.

Allow-list filters from the generated field methods above. For reverse relations, use only the
generated relation selector and its `With(childRequest)` variant listed above;
use generated relation loading rather than handwritten child-query loops; never
infer child names.

`Purpose(...)` changes the request into its executable stage. `Comment(...)` may be supplied before or after it, but execution requires both values. `ExecuteForList` has exactly one context argument: `context`; services and trusted policy are injected when that `UserContext` is initialized.

## Field-specific Query Assist

Use the canonical KSML field name from this list. Do not substitute a language member name, JSON name, or database column.

| KSML field | Type | Field help |
| --- | --- | --- |
| `id` | `id` | `golang-assist-query/order_item.id` |
| `customer_order` | `relation<customer_order>` | `golang-assist-query/order_item.customer_order` |
| `name` | `string` | `golang-assist-query/order_item.name` |
| `version` | `version` | `golang-assist-query/order_item.version` |


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

Capability: `query`.

- Use generated projection, relation, predicate, sort, and aggregate APIs; never
  concatenate application SQL or silently ignore an unsupported filter.
- Rows, record count, facets, and aggregates must share the same active filter.
- Add negative tests for missing purpose/comment and forbidden dynamic fields.

### Optional per-parent Top-N optimization

- Configure `TopNProbeParentThreshold(threshold)` on the nested child request before its intent/execution
  stage. Use the field-specific relation Assist for the exact generated selector.
- Give the child a per-parent limit and deterministic ordering with an ID tie-breaker;
  bound the parent query separately. A single-parent detail query is not batch Top-N.
- Server providers default to a window query (`threshold = 0`). A positive threshold
  permits bounded probes only when the already-loaded parent count is at or below it;
  above it, use the window plan. SQLite uses its internal AlwaysProbe policy by default.
- Measure before opting in. Do not infer a threshold from table size or execute COUNT
  or statistics queries to choose a plan. Keep an index on foreign key + sort + ID.
- Inspect selected-plan, parent-count, per-parent-limit and probe-count telemetry.
  Runtime-managed probes are not application-owned N+1 loops; neither strategy changes
  authorization, loaded-state semantics or the required comment/purpose.
