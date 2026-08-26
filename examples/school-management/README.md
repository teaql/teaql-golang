# School Management bootstrap example

This generated example retains the shared `models/school-model.xml` fixture and
verifies SQLite bootstrap semantics against the local runtime:

- `ensureSchema` is explicit;
- Platform root `id=1` and SchoolType constants `1001`/`1002` are seeded;
- an unchanged second ensure is idempotent and keeps `version=1`;
- changing a constant in the module reconciles it once and increments its version.

Run `go test ./...` from this directory. The local `replace` is intentional for
pre-publication runtime verification.
