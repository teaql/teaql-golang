# School startup and audited directory showcase

The model is `models/school-model.xml`. `golang-lib-core` contains the exact
generated domain library; `golang-app-console` is the generated workspace.
Its console initializes the platform/types, registers a real School and reads
it back. Repeating the console reuses that School. All three original business
test files live in the workspace; only their package/import boundary changed.
`showcase.json` records the model, scope and producer identities.

This generated example retains the shared `models/school-model.xml` fixture and
verifies SQLite bootstrap semantics against the local runtime:

- `ensureSchema` is explicit;
- Platform root `id=1` and SchoolType constants `1001`/`1002` are seeded;
- an unchanged second ensure is idempotent and keeps `version=1`;
- changing a constant in the module reconciles it once and increments its version.

The generated library was regenerated as a complete artifact from this retained
model and the paired local generator, not manually patched. The application-owned
`golang-app-console/request_intent_test.go` also verifies the required request contract with SQL
logging disabled:

- blank Query comment or purpose is rejected on list, page and stream paths;
- Save without `AuditAs` is rejected and creates no School row;
- repeated ensure retains Platform and both fixed-ID SchoolType constants.

Run `bash verify.sh` from this directory. It binds the generated library,
workspace and this runtime in a disposable Go workspace. It runs all retained
tests twice with the race detector, runs the console twice on one retained
SQLite file, checks results through a separate read-only SQLite connection and
removes its owned build outputs. The library manifest remains generated. Individual tests use
temporary SQLite paths; two suite invocations are not a no-cleanup graph replay
or proof of full hierarchical Trace Chain completion.

## Generated bootstrap Trace Chain gate

`bash scripts/verify-school-bootstrap-example.sh` (from the runtime root) runs
the actual generated School bootstrap twice per retained SQLite file, with SQL
logs on/off and the race detector enabled. No trace nodes are injected. It
checks request intent, canonical physical paths, matching typed mutation and
audit lineage, caller identity restoration, no-op reseeding and an audited
constant edit/reconciliation. An independent read-only connection must see the
expected version at audit delivery. Generated-library hashes must stay unchanged.
