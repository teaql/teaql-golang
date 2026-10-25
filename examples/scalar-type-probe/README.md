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

Time is `core.TimeOfDay`, or `*core.TimeOfDay` when optional. It is neither a
timestamp nor a duration; zero means midnight. `core.ParseTimeOfDay` validates
`HH:mm:ss` with up to nine fractional digits and no timezone/date. SQLite stores
all nanoseconds exactly. PostgreSQL/MySQL binders explicitly reject unsupported
sub-microsecond values; MySQL uses TIME(6). These provider binding tests are not
claims of live PostgreSQL/MySQL integration acceptance.

The example verifies typed setter → SQLite → time predicate/projection → E,
null/NotLoaded, and immediate E access to an unsaved native Timestamp. A native
clock value also round-trips through its text JSON representation without I/O.
This is **not typed entity JSON** or a query/facet response envelope.

One clearly labelled raw-SQL fault injection corrupts a separate test row.
Generated Q hydration must reject it at `required_time`, never silently return
midnight. Read-only physical SQL is an independent storage oracle, not the
business query implementation. Existing noncanonical string-Time columns/data
need an explicit migration review; the example creates isolated fresh databases
and does not silently migrate legacy storage.

`generated.sha256` and `generated-provenance.json` identify the evaluated model,
producer and byte-exact generated files. Regenerate from the model; do not edit
the library. Discover API changes through current entity/action and field Assist.
