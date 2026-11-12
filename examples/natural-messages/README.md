# Natural Checker messages through a generated SQLite graph

Run `node examples/natural-messages/verify-local.cjs`. The example compiles with
this checkout, not a released runtime, and uses the unchanged generated library
and model under `../context-derived/fixture`. It runs twice against one isolated
SQLite database. Use `TEAQL_NATURAL_MESSAGES_EVIDENCE_DIR` with a new path to
choose where compact evidence and the database are retained.

Generated Q/E and audited graph saves demonstrate English and Chinese required
messages for the second child, safe original scalar-length reporting for a
sensitive field, no provider activity on rejection, and one persisted/audited
child update after correction. Application-owned metadata supplies presentation
labels, a name sensitivity policy and an eight-scalar example constraint. This
does not claim automatic generator emission of the new presentation metadata.

For shared golden cases, all 270 Locale/rule/depth templates, the native root
module suite and runtime race tests, run:

```bash
node scripts/verify-natural-messages.cjs --full --example --output /new/evidence/path
```

The verifier retains commands, origins, source/fixture hashes, responses and
outcomes. Disposable executables and scratch directories are removed after
their owned children exit. Shared Go build and module caches are not deleted.
