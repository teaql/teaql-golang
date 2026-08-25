# Go retained minimum conformance example

This generated SQLite workspace retains executable evidence for the minimum
TeaQL runtime contract:

1. explicit, idempotent `EnsureSchema`;
2. checker rejection before persistence;
3. persisted strong-type return from create;
4. typed `SmartList` query return;
5. E-expression loaded, null, and not-loaded states;
6. optimistic version advancement on update; and
7. soft-delete visibility rules.

Run it from this directory:

```bash
go test ./...
go run ./src
```

The generated domain library is under `lib/`; `src/conformance_test.go` is the
retained executable specification. Installing a runtime module does not mutate
schema. The application calls `lib.EnsureSchema(context)` explicitly.

Root/constant seed reconciliation is tracked separately as a generator gap and
is intentionally not reported as passing by this example.
