# School bootstrap and repeated-bootstrap evidence

Date: 2026-08-26  
Runtime revision: `4d6c677`  
Generator endpoint: `https://api.teaql.io/latest/generate`  
Generator target: `golang-lib-core`  
Database: SQLite  
Tracking: [teaql-golang#5](https://github.com/teaql/teaql-golang/issues/5)

## Contract under test

Using [school-bootstrap-model.xml](../school-bootstrap-model.xml):

1. The first explicit `EnsureSchema` creates the schema, Platform id `1`, and SchoolType constants `1001` and `1002`.
2. The second explicit `EnsureSchema` is idempotent: it retains exactly one root and two constants.
3. Bootstrap reconciliation advances the SchoolType ID floor beyond `1002`.

Installing the Runtime Module must remain passive.

## Execution

The generated workspace was created with:

```bash
cargo teaql --input school-bootstrap-model.xml \
  --output /tmp/teaql-school-go golang-lib-core
```

A fresh SQLite database was used. The generated `ServiceRuntimeFromEnv()` was started, `EnsureSchema(context)` was called twice, and both tables were counted after each call.

## Actual result

```text
first platform=0 constants=0
second platform=0 constants=0
```

The process exited successfully: both DDL passes completed, but no bootstrap records were written.

## Result

| Assertion | Result |
| --- | :---: |
| First explicit schema reconciliation succeeds | PASS |
| Second identical reconciliation succeeds | PASS |
| Platform id 1 is seeded | FAIL |
| SchoolType 1001/1002 are seeded | FAIL |
| Repeated bootstrap retains 1 root / 2 constants | FAIL |
| Bootstrap advances constant ID floor | NOT REACHED |

Overall: **GAP**. This is not equivalent to the separately passing explicit-ID/ID-floor unit test. Generated root and constant metadata must be added to the Runtime Module and reconciled explicitly by `EnsureSchema`.

