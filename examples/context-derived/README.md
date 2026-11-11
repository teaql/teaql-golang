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
