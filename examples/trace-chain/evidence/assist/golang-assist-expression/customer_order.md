<!-- ephemeral -->

# Go Assist — Expression `Customer Order`

The generated Expression facade preserves Value, loaded null, and NotLoaded.
`Eval()` returns `(value, present)` for the first two and panics with
`TeaQLNotLoadedError` for the third. `OrElse` is null-only and never hides
NotLoaded.

```go
package assist_expression

import "trace-chain-service-core-workspace/lib/customer_order"

func ExtractCustomerOrderId(entity *customer_order.CustomerOrder) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).Id().Eval()
}

func ExtractCustomerOrderIdOrElse(entity *customer_order.CustomerOrder, fallback uint64) uint64 {
	return customer_order.NewCustomerOrderExpression(entity).Id().OrElse(fallback)
}

func ExtractCustomerOrderOrderNumber(entity *customer_order.CustomerOrder) (string, bool) {
	return customer_order.NewCustomerOrderExpression(entity).OrderNumber().Eval()
}

func ExtractCustomerOrderOrderNumberOrElse(entity *customer_order.CustomerOrder, fallback string) string {
	return customer_order.NewCustomerOrderExpression(entity).OrderNumber().OrElse(fallback)
}

func ExtractCustomerOrderDescription(entity *customer_order.CustomerOrder) (string, bool) {
	return customer_order.NewCustomerOrderExpression(entity).Description().Eval()
}

func ExtractCustomerOrderDescriptionOrElse(entity *customer_order.CustomerOrder, fallback string) string {
	return customer_order.NewCustomerOrderExpression(entity).Description().OrElse(fallback)
}

func ExtractCustomerOrderVersion(entity *customer_order.CustomerOrder) (int64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).Version().Eval()
}

func ExtractCustomerOrderVersionOrElse(entity *customer_order.CustomerOrder, fallback int64) int64 {
	return customer_order.NewCustomerOrderExpression(entity).Version().OrElse(fallback)
}

func ExtractCustomerOrderPlatformId(entity *customer_order.CustomerOrder) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).PlatformId().Eval()
}

func TraverseCustomerOrderPlatform(entity *customer_order.CustomerOrder) *customer_order.PlatformRelationExpression {
	return customer_order.NewCustomerOrderExpression(entity).Platform()
}

func AggregateCustomerOrderOrderItemListSize(entity *customer_order.CustomerOrder) (int, bool) {
	return customer_order.NewCustomerOrderExpression(entity).OrderItemList().Size().Eval()
}

func FirstCustomerOrderOrderItemListId(entity *customer_order.CustomerOrder) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).OrderItemList().First().Id().Eval()
}

func GetCustomerOrderOrderItemListId(entity *customer_order.CustomerOrder, index int) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).OrderItemList().Get(index).Id().Eval()
}

func AggregateCustomerOrderPaymentListSize(entity *customer_order.CustomerOrder) (int, bool) {
	return customer_order.NewCustomerOrderExpression(entity).PaymentList().Size().Eval()
}

func FirstCustomerOrderPaymentListId(entity *customer_order.CustomerOrder) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).PaymentList().First().Id().Eval()
}

func GetCustomerOrderPaymentListId(entity *customer_order.CustomerOrder, index int) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).PaymentList().Get(index).Id().Eval()
}

func AggregateCustomerOrderShipmentListSize(entity *customer_order.CustomerOrder) (int, bool) {
	return customer_order.NewCustomerOrderExpression(entity).ShipmentList().Size().Eval()
}

func FirstCustomerOrderShipmentListId(entity *customer_order.CustomerOrder) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).ShipmentList().First().Id().Eval()
}

func GetCustomerOrderShipmentListId(entity *customer_order.CustomerOrder, index int) (uint64, bool) {
	return customer_order.NewCustomerOrderExpression(entity).ShipmentList().Get(index).Id().Eval()
}


```

Select every traversed field and relation. Never recover a
`TeaQLNotLoadedError` merely to supply a default, and never replace generated
Expression accessors with direct zero-value getters.

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

Capability: `expression`.

- Distinguish a loaded null from a field or relation that was not projected. A
  NotLoaded/coding error must remain visible; do not turn it into an ordinary null.
- Select every traversed relation first and use the generated E/expression API for
  scalar, object, and list traversal. Do not translate Java accessor names by guess.
