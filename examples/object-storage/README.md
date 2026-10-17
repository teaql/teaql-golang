# Context owned object storage

Tracking: [Go #50](https://github.com/teaql/teaql-golang/issues/50),
[shared contract #122](https://github.com/teaql/teaql-conformance/issues/122).

The core values and context facade retain the runtime's Go 1.18 baseline.
The optional `github.com/teaql/teaql-golang/objectstorage/s3` module uses the
official AWS SDK v2 and needs Go 1.24+. Trusted startup configures one shared
SDK client, bucket, credentials/refresh, endpoint and timeouts. Applications
own the HTTP pool and close idle connections at shutdown.

```go
provider, err := storages3.New(sdkClient, "attachments")
context := runtime.NewUserContext().WithObjectStorage(provider)
storage, err := context.ObjectStorage("store payment receipt attachment")
key, err := core.NewObjectKey("receipts/opaque-reference.pdf")
options, err := core.NewObjectPutOptions(size, core.ObjectCreateOnly)
metadata, err := storage.Put(key, source, options)
link, err := storage.GetReadURL(key, core.DefaultObjectReadURLOptions())
// Return link.URL() only to an authorized caller; do not log it.
```

Upload inputs are closed exactly once. Uploads stage in a private temporary file
using bounded buffers, exact length validation and a 5 GiB single-PUT limit.
Downloads stream; close them explicitly. `GetBytes` requires a maximum size.
Create-only uses conditional PUT, replacement is explicit, and denied/unavailable
responses never become false absence. Keys and temporary reusable bearer URLs
are distinct. The default URL lifetime is five minutes; credentials/policy may
expire it earlier. URL issuance does not check existence or infer permissions.
Diagnostic formatting and operation events exclude private storage values and
raw SDK errors. Object storage effects remain outside SQL transactions.

Set `TEAQL_S3_ENDPOINT`, `TEAQL_S3_BUCKET`, `TEAQL_S3_ACCESS_KEY` and
`TEAQL_S3_SECRET_KEY` for a dedicated test service, then run:

```bash
bash examples/object-storage/verify.sh
bash examples/object-storage/verify.sh
```

Set `GO=/path/to/new/go` if the system compiler predates Go 1.24. The example
resolves both runtime and S3 module through explicit local replacements. It
provisions the test bucket if absent and deletes only its unique test keys;
the provider performs no bucket administration. The gate covers offline contracts,
actual SDK HTTP failures and a live lifecycle. AWS and other compatible servers
need separate qualification. Multipart/list/copy/range/signed-upload are extensions.
