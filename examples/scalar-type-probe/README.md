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
