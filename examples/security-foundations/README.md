# Security foundations example

This executable uses the runtime source in this repository and fails closed if
any of these backend security contracts regress:

1. the ordinary default-on SQL log is parameterized and value-free, while an
   explicitly installed sensitive sink receives copy-paste SQL;
2. the trusted TFP endpoint rejects tenant/raw-SQL injection and carries the
   trusted tenant, ID, and optimistic version into one mutation statement; and
3. a `UserContext` opaque entity reference matches the shared Go/.NET
   AES-256-GCM golden vector and rejects a wrong purpose.

Run it from the repository root:

```bash
go run ./examples/security-foundations
```

This example does not enable the development-only raw-ID escape hatch and does
not replace authentication or application authorization.
