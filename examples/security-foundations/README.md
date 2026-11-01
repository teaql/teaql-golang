# Security foundations example

This executable uses the runtime source in this repository and fails closed if
any of these backend security contracts regress:

1. the ordinary default-on SQL log contains copy-paste SQL with masked values;
   a sensitive sink still needs the explicit plaintext-debug acknowledgement;
2. the trusted TFP endpoint rejects tenant/raw-SQL injection and carries the
   trusted tenant, ID, and optimistic version into one mutation statement; and
3. a `UserContext` opaque entity reference matches the shared Go/.NET
   AES-256-GCM golden vector and rejects a wrong purpose; and
4. a fixed, context-owned Business Clock drives both business time and business
   date without changing any process-global clock; and
5. Query Policy operates on an execution snapshot, adds trusted scope exactly
   once, leaves the caller request reusable, and fails closed on denial; and
6. future or unreadable Business Clock values cannot alter opaque-reference
   issue/expiry time. The example calibrates the unreadable clock, then proves
   security work makes zero calls to it and uses the real operational clock.

Run it from the repository root:

```bash
go run ./examples/security-foundations
```

This example does not enable the development-only raw-ID escape hatch and does
not replace authentication or application authorization.

The public-API regression tests in
`runtime/operational_reference_clock_test.go` additionally exercise forward and
backward jumps, actual deadline expiry while business time stays frozen, wrong
type/purpose and tamper rejection, and both explicitly enabled raw-reference
paths. They do not substitute Business Clock for the codec's separate security
clock SPI, nor claim cache/HTTP timeout coverage.
