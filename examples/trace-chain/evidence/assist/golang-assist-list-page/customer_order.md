<!-- ephemeral -->

Please help me complete the pagination list (List Page) service business code for the `Customer Order` object.

### Standard List Page Example (Reference)
```go
package assist_list_page

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	. "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/payment"
	"trace-chain-service-core-workspace/lib/shipment"

)

func ListPageExample(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*customer_order.CustomerOrder], error) {
	listResult, err := Q.CustomerOrdersMinimal().
		SelectId().
		SelectPlatform().
		SelectOrderNumber().
		SelectDescription().
		SelectVersion().
		SelectOrderItemListWith(order_item.NewOrderItemMinimalRequest().Limit(20)).
		SelectPaymentListWith(payment.NewPaymentMinimalRequest().Limit(20)).
		SelectShipmentListWith(shipment.NewShipmentMinimalRequest().Limit(20)).
		OrderByIdAsc().
		Comment("what: Retrieve paginated list").
		Purpose("why: Paginate entities").
		ExecuteForPage(context, offset, size)

	if err != nil {
		return nil, err
	}

	return listResult, nil
}
```

### Your Task
Please implement the real list-page logic for `Customer Order` based on my specific business needs.

Validate `offset` at the application boundary. TeaQL rejects zero-sized pages and applies a fixed runtime list hard limit of 10,000 that application and federated callers cannot override.
Use a stable unique generated order, allow-list field filters/sorts, and bound every selected
reverse relation. `ExecuteForPage` computes the exact count from the same authorized filters while excluding projection, relations, ordering, and pagination.

### Bounded Entity Streaming

```go
func StreamExample(context *runtime.UserContext, yield func(*customer_order.CustomerOrder) error) error {
	return Q.CustomerOrdersMinimal().
		OrderByIdAsc().Limit(20).
		Comment("what: consume a bounded root-entity stream").
		Purpose("why: process projected entities without buffering the whole result").
		ExecuteForStream(context, 10, yield)
}
```

The provider must implement `StreamQueryExecutor`. Returning an error from the
consumer stops iteration and releases the cursor. This path delivers projected
root entities, not hydrated relation graphs; use list/page execution when E must
traverse selected relations.

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

Capability: `list-page`.

- Validate offset, page size, filters, deep paths, IN-list size, and sort against
  explicit allow-lists. Reject invalid input instead of widening the query.
- Use a stable unique ordering and retain the runtime hard limit. Continuous-page
  optimization is opt-in, browsing-only, local runtime policy and cannot cross TFP.
- Run count only when explicitly requested; otherwise use the returned list length.
