<!-- ephemeral -->

# Go Assist — Expression `Platform`

The generated Expression facade preserves Value, loaded null, and NotLoaded.
`Eval()` returns `(value, present)` for the first two and panics with
`TeaQLNotLoadedError` for the third. `OrElse` is null-only and never hides
NotLoaded.

A selected forward relation expression also supports `Eval()`, returning
`(core.Record, present)` and preserving the same NotLoaded guard. If a nested
request selected an explicitly named related aggregate, read that alias from
the returned record and use the runtime Value conversion (e.g. `TryU64()` for a
count). Do not treat an absent alias as zero. Discover the aggregate selector
through the reverse-relation field Query Assist.

```go
package assist_expression

import "trace-chain-service-core-workspace/lib/platform"

func ExtractPlatformId(entity *platform.Platform) (uint64, bool) {
	return platform.NewPlatformExpression(entity).Id().Eval()
}

func ExtractPlatformIdOrElse(entity *platform.Platform, fallback uint64) uint64 {
	return platform.NewPlatformExpression(entity).Id().OrElse(fallback)
}

func ExtractPlatformName(entity *platform.Platform) (string, bool) {
	return platform.NewPlatformExpression(entity).Name().Eval()
}

func ExtractPlatformNameOrElse(entity *platform.Platform, fallback string) string {
	return platform.NewPlatformExpression(entity).Name().OrElse(fallback)
}

func ExtractPlatformVersion(entity *platform.Platform) (int64, bool) {
	return platform.NewPlatformExpression(entity).Version().Eval()
}

func ExtractPlatformVersionOrElse(entity *platform.Platform, fallback int64) int64 {
	return platform.NewPlatformExpression(entity).Version().OrElse(fallback)
}

func AggregatePlatformCustomerOrderListSize(entity *platform.Platform) (int, bool) {
	return platform.NewPlatformExpression(entity).CustomerOrderList().Size().Eval()
}

func FirstPlatformCustomerOrderListId(entity *platform.Platform) (uint64, bool) {
	return platform.NewPlatformExpression(entity).CustomerOrderList().First().Id().Eval()
}

func GetPlatformCustomerOrderListId(entity *platform.Platform, index int) (uint64, bool) {
	return platform.NewPlatformExpression(entity).CustomerOrderList().Get(index).Id().Eval()
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
