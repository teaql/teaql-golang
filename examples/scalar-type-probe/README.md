# Generated scalar and typed wall-clock example

From the runtime repository root:

```bash
bash examples/scalar-type-probe/verify.sh
```

The script binds this repository, the application and its retained generated
library in an owned Go consumer with explicit local replacements. It does not
use an older published runtime, edit generated source, or rely on an ambient
workspace. It compiles once and executes the binary twice without deleting
its SQLite files. Build/workspace outputs are
removed on exit; the printed evidence directory keeps logs and databases.
Set `GO` to an explicit compatible compiler if the system compiler is old.

The unchanged evaluated model exercises 21 business scalar fields with generated
Q/E/audited-save APIs. Exact Decimal and Long values, false/empty values, nullable
fields and projection NotLoaded remain distinct. Twelve nonfinite mutation cases
must fail before query, mutation, transaction begin or ID allocation. Successful
provider operations calibrate those counters.

## Boolean admission and readback

The default verifier also uses generated Q/E/audited saves for true, false,
explicit null, projection NotLoaded and optimistic versions 1→2→3. Twelve
create/update custom-Fix attempts replace required/optional Booleans with
integers, text or JSON bytes. All must fail before query, mutation, transaction
begin or ID allocation, without changing pending values, identity, status or
original version. A valid custom repair still saves false. One explicitly
labelled SQL fault injection tests corrupt storage, then restores the control.
Two executions preserve four distinct rows in `scalars.sqlite.boolean.sqlite`.

Native transport cases additionally cover INTEGER, BOOL and legacy BOOLEAN
columns, ordinary/named/grouped results, streaming, renamed columns, numeric
COUNT aliases and Boolean MIN/MAX. SQLite result expressions retain physical
values instead of accepting the driver's early nonzero-to-true conversion;
predicates, grouping and ordering keep their original column affinity. This
is modeled query/mutation coverage, not arbitrary raw-SQL result typing or
full scalar/provider/release acceptance.

## Exact integral admission after custom Fix

```bash
node examples/scalar-type-probe/verify-integral.cjs
```

This focused companion compiles once with the race detector and runs twice on
one SQLite database without reset. Fifteen invalid modeled-Integer/Long/reference
values are tested on both create and update. A custom Fix injects fractional
Decimals, overflow, negative unsigned references, floats, text and booleans.
Each rejection must identify the field with `invalid_type` before query,
mutation, transaction begin or ID allocation. One successful generated save
and Q reload calibrate all observers; E and read-only SQLite confirm unchanged
Long, FK and optimistic version. Decimal-to-integer conversion must be exact,
including extreme exponent rejection without allocating a huge power of ten.

The generated model marks Integer as a signed 32-bit domain while retaining
the native int64 accessor and SQLite INTEGER column. Both normal generated
Checker and custom Fix paths reject out-of-range required/optional Integers
before I/O. Eight additional native-updater cases cover both bounds on create
and update; legal endpoints round-trip through Q/E and audited version 1→2.
Long remains 64-bit. The fixture reuses the unchanged evaluated model and a
freshly generated library. This is mutation admission, not qualification of
business ID positivity, full canonical relation
naming, live external databases or an installed-package release.

## Invalid legacy Integer reads

The main `verify.sh` also runs 40 read-side cases per round: eight real SQLite
corruptions, eight failed `FromRecord` calls with entity/ledger atomicity, and
24 E terminals (`TryEval`, `Eval`, `OrElse`). Required and optional Integer
fields reject overflow, underflow, fractional REAL and nonnumeric text. Errors
identify the canonical field/entity/expected type without retaining raw values.
A failed read does not write or allocate IDs. Safe projections still leave
unselected fields NotLoaded. `OrElse` must not disguise an invalid loaded value;
genuine loaded null and ordinary NotLoaded semantics remain unchanged.

SQL and Base writes in this test are explicitly labelled fault injections.
Generated audited saves create the rows, generated Q performs the tested reads,
and generated E supplies the values/errors. Two executions preserve 18 rows
(16 intentionally corrupt) at version 1; the original valid Integer/Long control
remains readable. This does not qualify every scalar family or external database.

Time is `core.TimeOfDay`, or `*core.TimeOfDay` when optional. It is neither a
timestamp nor a duration; zero means midnight. `core.ParseTimeOfDay` validates
`HH:mm:ss` with up to nine fractional digits and no timezone/date. SQLite stores
all nanoseconds exactly. PostgreSQL/MySQL binders explicitly reject unsupported
sub-microsecond values; MySQL uses TIME(6). These provider binding tests are not
claims of live PostgreSQL/MySQL integration acceptance.

The example verifies typed setter → SQLite → time predicate/projection → E,
null/NotLoaded, and immediate E access to an unsaved native Timestamp. A native
clock value also round-trips through its text JSON representation without I/O.
The separate entity JSON test now also exercises the actual generated codec.

## Typed entity JSON

`entity.ToJSON()` and `encoding/json.Marshal(entity)` use model-owned metadata.
`lib.ScalarTypeProbeFromJSON(payload)` returns the actual generated entity, with
independent clean mutation ownership. A default camelCase model can explicitly
use `runtime.JsonFieldSnakeCase` or `runtime.JsonFieldPascalCase` in both calls.
The metadata is in an internal leaf package; concrete factories belong to the
library facade, avoiding parent/child package import cycles and global registries.

Each execution verifies three profiles, all 21 business scalar fields, selected
forward and reverse relations, exact Long/Decimal and nanosecond Time, omitted
versus explicit null, zero-I/O JSON/E, rejection of partial decoded mutations,
save of a complete decoded entity and rejection of an independently decoded
stale version. The two executions keep the same database and preserve six
distinct JSON entity IDs. Native provider and codec tests additionally cover
typed SQL NULL, identity nulls, duplicate keys, aliases, UTF text, cycles and bounds.

JSON is a loaded snapshot, not a create command or authorization decision.
Audited save and Checker still apply. This is not a query/facet response envelope
or a context-bound external-reference transport adapter.

One clearly labelled raw-SQL fault injection corrupts a separate test row.
Generated Q hydration must reject it at `required_time`, never silently return
midnight. Read-only physical SQL is an independent storage oracle, not the
business query implementation. Existing noncanonical string-Time columns/data
need an explicit migration review; the example creates isolated fresh databases
and does not silently migrate legacy storage.

`generated.sha256` and `generated-provenance.json` identify the evaluated model,
producer and byte-exact generated files. Regenerate from the model; do not edit
the library. Discover API changes through current entity/action and field Assist.
