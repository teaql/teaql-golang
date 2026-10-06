<!-- ephemeral -->

# Go Assist — Update `Order Item`

Load the current row and original optimistic version. Never reconstruct an
entity from an ID/version supplied by a client.

An entity update is not a partial DTO patch. Load every scalar field before
modification so checker rules can validate the complete business state. Never
use the `Minimal()` request for an entity that will be saved.

```go
package assistupdate

import (
    "fmt"

    "github.com/teaql/teaql-golang/runtime"
    . "trace-chain-service-core-workspace/lib"
    "trace-chain-service-core-workspace/lib/order_item"
)

func UpdateOrderItem(
    context *runtime.UserContext,
    id uint64,
    name string,
    failIfMissing bool,
) (*order_item.OrderItem, error) {
    entity, err := Q.OrderItems().
        WithIdIs(id).

        Comment("what: load current Order Item for update").
        Purpose("why: preserve original version for optimistic locking").
        ExecuteForOne(context)
    if err != nil {
        return nil, err
    }
    if entity == nil {
        if failIfMissing {
            return nil, fmt.Errorf("Order Item not found: %d", id)
        }
        return nil, nil
    }

    entity.UpdateName(name)

    saved, err := entity.
        AuditAs("Update Order Item for the requested business operation").
        Save(context)
    if err != nil {
        return nil, err
    }
    return saved, nil
}
```

Compile and execute this source unchanged. Prove persistence and query-back, a
conflict from an independently loaded stale copy, nil and error not-found
behavior, missing/blank audit rejection, and compilation failure for unknown or
trusted fields. Constant and relation changes require an explicitly selected
generated method; do not invent one.

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

Capability: `update`.

- Load the tenant-scoped current entity first so its original version participates
  in optimistic locking; do not reconstruct versioned state from untrusted JSON.
- Allow-list writable fields, attach the generated audit-reason API, and save with
  the same UserContext. Add a stale-version rejection test.
