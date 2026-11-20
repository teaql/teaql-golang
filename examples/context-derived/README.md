# Context-derived fields without synthetic mutations

Run `node examples/context-derived/verify-local.cjs` from this runtime checkout.
The verifier binds the generated library to this repository, compiles a disposable
race-enabled consumer, then runs it twice against the same isolated SQLite database.
Set `TEAQL_CONTEXT_FIX_EVIDENCE_DIR` to a new directory to retain compact logs,
hashes, dependency origins and the test database. Build outputs are removed on exit.

The retained model and generated files are immutable fixtures. Regenerate them
with the generator's `GolangContextFixGenerationTest` and
`-Dteaql.go.context.fixture.output=<absolute fixture directory>`; do not patch them.

Covered through generated Q/E and audited save:

- A different clock, IP and actor cannot turn two unchanged saves into an update.
- Real updates receive update-scoped Context values and preserve create-only ones.
- Validation/provider failures report ObjectLocation and preserve caller edits.
- A later invalid child rolls back an earlier sibling's staged Context fields.
- Retry changes only the independently modified child, not its parent or sibling.

IP/actor ownership is an application metadata/Checker SPI example, not a claim
that new built-in KSML IP or actor field types have been introduced.

## Current business model and workspace

`model.xml` defines Catalog and Entry. Trusted Context supplies collection/edit
times, IP and actor fields; the example shows why saving an unchanged entry
must leave those business values untouched. `golang-lib-core/` is the exact, read-only
current generated domain library. The application is in `golang-app-console/context_fix_test.go`.
The historical `fixture/` remains byte-identical and retains all of its original
fingerprint checks; a separate guard verifies the current library before and
after business execution.

Run the focused source qualifier entry from the SDK repository root:

```bash
node examples/context-derived/verify-local.cjs
```

The original assertions run twice on retained stores: no-op saves produce no
updates, genuine edits consume the trusted Context, failed validation/provider
work preserves edits, and a later invalid child rolls back staged graph fields.
Current managed creation results are explicitly awaited or checked. No generated
implementation is hand-edited. Compact receipts and result stores are retained;
owned build output is removed. Full runtime inventories remain separate gates.

The current model adds only `_audit_mask_fields="guarded_token"` to satisfy the
current evaluator. Business fields/defaults and the frozen model are unchanged.
