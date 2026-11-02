# SQLite connection lifecycle

This example saves and reloads a small generated PurchaseOrder and OrderItem
graph after physical SQLite connections change. Every new connection must retain
Soundex, exact Decimal validation, exact ordering and exact SUM. Empty SUM stays
NULL; values larger than binary floating point can represent exactly are checked
by separate scalar queries.

Run with Node.js, a CGO-enabled Go toolchain and a C compiler:

```bash
node examples/sqlite-lifecycle/verify-local.cjs
```

The verifier binds the runtime to this checkout and uses the retained generated
library and model in [fix-e-alignment](../fix-e-alignment). It runs the same binary
twice without resetting its four databases. It checks physical connection
discard, three simultaneously leased connections, idle eviction and cancellation
of an owned SQL transaction. The application UserContext is never replaced.
All application writes use generated audited saves; readback uses generated Q/E.

For manual SQLite setup, create the pool with
`sqlite.OpenDatabase(dataSourceName)`. Its driver initializes every connection;
opening a pool does not create schema. The generated runtime uses this factory
and still requires the normal Context-based `EnsureSchema` call. The default
pool limit is one, supporting `:memory:`; file-backed applications can explicitly
increase it. The legacy `EnsureSoundex` initializer alone cannot retrofit hooks
onto future connections of an arbitrary `sql.Open("sqlite3", ...)` pool.

Temporary builds and default evidence are removed on exit. To retain compact
logs, reports and databases, supply a new absolute output directory:

```bash
TEAQL_SQLITE_LIFECYCLE_EVIDENCE_DIR=/absolute/new-evidence-directory \
  node examples/sqlite-lifecycle/verify-local.cjs
```

The full repository examples gate also runs this verifier. No external database,
package publication or generator server is needed to run the retained example.
