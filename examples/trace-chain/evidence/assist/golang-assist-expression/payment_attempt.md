<!-- ephemeral -->

# Go Assist — Expression `Payment Attempt`

The generated Expression facade preserves Value, loaded null, and NotLoaded.
`Eval()` returns `(value, present)` for the first two and panics with
`TeaQLNotLoadedError` for the third. `OrElse` is null-only and never hides
NotLoaded.

```go
package assist_expression

import "trace-chain-service-core-workspace/lib/payment_attempt"

func ExtractPaymentAttemptId(entity *payment_attempt.PaymentAttempt) (uint64, bool) {
	return payment_attempt.NewPaymentAttemptExpression(entity).Id().Eval()
}

func ExtractPaymentAttemptIdOrElse(entity *payment_attempt.PaymentAttempt, fallback uint64) uint64 {
	return payment_attempt.NewPaymentAttemptExpression(entity).Id().OrElse(fallback)
}

func ExtractPaymentAttemptReferenceCode(entity *payment_attempt.PaymentAttempt) (string, bool) {
	return payment_attempt.NewPaymentAttemptExpression(entity).ReferenceCode().Eval()
}

func ExtractPaymentAttemptReferenceCodeOrElse(entity *payment_attempt.PaymentAttempt, fallback string) string {
	return payment_attempt.NewPaymentAttemptExpression(entity).ReferenceCode().OrElse(fallback)
}

func ExtractPaymentAttemptVersion(entity *payment_attempt.PaymentAttempt) (int64, bool) {
	return payment_attempt.NewPaymentAttemptExpression(entity).Version().Eval()
}

func ExtractPaymentAttemptVersionOrElse(entity *payment_attempt.PaymentAttempt, fallback int64) int64 {
	return payment_attempt.NewPaymentAttemptExpression(entity).Version().OrElse(fallback)
}

func ExtractPaymentAttemptPaymentId(entity *payment_attempt.PaymentAttempt) (uint64, bool) {
	return payment_attempt.NewPaymentAttemptExpression(entity).PaymentId().Eval()
}

func TraversePaymentAttemptPayment(entity *payment_attempt.PaymentAttempt) *payment_attempt.PaymentRelationExpression {
	return payment_attempt.NewPaymentAttemptExpression(entity).Payment()
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
