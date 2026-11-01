# Generated E diagnostics under model JSON naming profiles

Run from this runtime repository:

```sh
node examples/expression-profiles/verify-local.cjs
```

The local `go.mod` is a module boundary, not a standalone consumer manifest.
Use the verifier above: it supplies the distinct generated dependency for each
profile. The runtime's root `go test ./...` must not try to compile this harness
without those dependencies; the complete example inventory runs it explicitly.

The same application executes against three evaluated, unmodified generated
libraries: default `camelCase`, `snake_case`, and `PascalCase`. Models differ
only in `json_field_naming`. The local producer identity and file hashes are
recorded in `provenance.json` and `generated/manifest.json`.

The application uses the context's schema bootstrap, a bounded generated Q
query, and the generated E API to traverse a reverse list which was not loaded.
The default error DTO must use the model's JSON naming, while preserving the
canonical KSML path, the full attempted native chain, one first-failure
`<broken>` marker, and preload guidance. It does not override the error profile.
An observed real Q query calibrates the provider counter; E causes no further
provider calls and leaves the selected root name unchanged.

Each model runs twice on the same SQLite database without a reset. Source,
fixture, command, origin and database evidence is retained; only task-owned
build scratch is removed. To choose an output location, set
`TEAQL_GENERATED_E_PROFILE_EVIDENCE_DIR` to a new absolute directory.
Existing reports and databases are never overwritten. Go resolves both the generated module and runtime through explicit local replacements, with public dependency downloads disabled.

By themselves these are local-source examples, not private/public package
qualification, a new audited-save test, a full inventory or full P0/P1 acceptance.
Literal punctuation and JSON-pointer escaping are covered by the separate
native expression-path tests; modeled names here are legal multiword fields.

## Complete inventory and copied-source origin

The repository inventory also runs all three modeled profiles twice. When an
outer gate compiles an isolated source copy, `TEAQL_GO_EXAMPLE_SOURCE_ROOT` names
its real Git origin. The verifier compares all source bytes, modes and link text
before compilation and records actual copied paths from `go list`; it does not
create Git aliases, rewrite generated files or substitute the origin path for
the code that was compiled.
