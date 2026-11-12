# Natural Checker messages through a generated SQLite graph

Run `node examples/natural-messages/verify-local.cjs`. The example compiles with
this checkout, not a released runtime, and uses the evaluated generated library
and model under `fixture/`. It runs twice against one isolated
SQLite database. Use `TEAQL_NATURAL_MESSAGES_EVIDENCE_DIR` with a new path to
choose where compact evidence and the database are retained.

Generated Q/E and audited graph saves demonstrate English and Chinese required
messages for the second child, safe original scalar-length reporting for a
sensitive field, no provider activity on rejection, and one persisted/audited
child update after correction. Both generated module factories install passive
presentation metadata. The application neither installs metadata nor replaces
Checkers, sensitivity or field constraints. The actual generated name limit is
100 Unicode scalars; a private 101-scalar input retains that original count.

Required string setters cannot accept nil. A separate new child omits its name
to exercise the generated required rule before identity reservation. Length
and customer-renderer failures use the persisted second child. All graph roots
remain unchanged; the corrected save changes only the intended second child.
The customer renderer receives masked arguments exactly once, and its safe
literal survives final graph relocation.

Go's internal entity key is `entry`, while the generated wire diagnostic type
is `Entry`. The runtime preserves that explicit diagnostic type and applies the
actual checked owner's privacy policy without guessing a casing conversion.

`fixture/generated-files.json` fingerprints all generated implementation;
`fixture/evaluation.json` records model evaluation, and `fixture/generation.json`
pins the upstream generator and generation receipt. Regenerate upstream rather
than editing generated files.

For shared golden cases, all 270 Locale/rule/depth templates, the native root
module suite and runtime race tests, run:

```bash
node scripts/verify-natural-messages.cjs --full --example --output /new/evidence/path
```

The verifier retains commands, origins, source/fixture hashes, responses and
outcomes. Disposable executables and scratch directories are removed after
their owned children exit. Shared Go build and module caches are not deleted.
