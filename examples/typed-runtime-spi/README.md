# Typed Runtime SPI example

Run `go run ./examples/typed-runtime-spi` from this repository. It installs a
schema executor independently of any query/mutation provider and verifies
precedence on two calls. It then installs a default and an application-owned
Internal ID generator on one `UserContext`, verifies explicit-ID floor
reservation, and proves a custom allocation error cannot fall through to the
default. The generator repository's school-model integration test separately
checks generated `EnsureSchema` and entity `Save` against a real SQLite database.
