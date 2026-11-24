# Purchase order service routing

The console installs a schema validator, a purchasing ID allocator and an
observable SQLite executor on the trusted Context. It places one order and
retries a revision after a write failure. It verifies rollback, custom identity,
and the last accepted revision after restart. A schema refusal skips business
execution. An allocation refusal skips mutation and the default generator.

`model.xml` defines the purchasing desk and its orders. The generated
`golang-app-console` uses Q and E from `golang-lib-core`. Run both launches
and the original typed SPI calibration with `bash examples/typed-runtime-spi/verify.sh`.

## Original schema and ID calibration

Run `go run ./examples/typed-runtime-spi` from this repository. It installs a
schema executor independently of any query/mutation provider and verifies
precedence on two calls. It then installs a default and an application-owned
Internal ID generator on one `UserContext`, verifies explicit-ID floor
reservation, and proves a custom allocation error cannot fall through to the
default. The generator repository's school-model integration test separately
checks generated `EnsureSchema` and entity `Save` against a real SQLite database.
