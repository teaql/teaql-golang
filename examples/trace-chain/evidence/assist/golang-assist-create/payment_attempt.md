<!-- ephemeral -->

# Go Assist — Create `Payment Attempt`

Use the exact generated `Q.PaymentAttempts()` entry point. The
trusted `runtime.UserContext` owns actor, tenant, policy, provider, entity
initialization, and audit infrastructure; none belongs in business input.

```go
package assistcreate

import (
	"github.com/teaql/teaql-golang/runtime"
	. "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/payment_attempt"
)

func CreatePaymentAttempt(
	paymentId uint64,
	referenceCode string,
	context *runtime.UserContext,
) (*payment_attempt.PaymentAttempt, error) {
	entity := Q.PaymentAttempts().
		Comment("what: initialize Payment Attempt").
		Purpose("why: create Payment Attempt").
		NewEntity(context)

	entity.UpdatePaymentId(paymentId)
	entity.UpdateReferenceCode(referenceCode)

	saved, err := entity.
		AuditAs("Create Payment Attempt for the requested business operation").
		Save(context)
	if err != nil {
		return nil, err
	}
	return saved, nil
}
```

Only the generated updater methods above are writable. Constant candidates,
when present, are generated under these exact public names:
Compile the source unchanged. Test persistence and query-back, and prove that
blank/missing intent, blank/missing audit reason, unknown fields, and attempted
trusted-context overrides fail. Do not edit generated sources.


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

Capability: `create`.

- Validate and allow-list writable business fields; never mass-assign dynamic JSON.
- Create through the generated request/entity API, attach a non-empty audit reason,
  save with the same UserContext, and return the runtime's native save result.
- Add a negative test proving a missing audit reason cannot write.
