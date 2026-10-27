# Optional Blob reference module

`NewImageContentVerifier` supplies the upload service's trusted verifier SPI for
`jpg`, `jpeg` and `png` using Go's standard JPEG and PNG decoders. The host creates
it with input, decoded-memory and concurrency limits, then passes it to
`NewSQLiteUploadService` before installing `BlobReferenceModule`. Construction is
passive; no parser or storage work starts until an upload is confirmed.

The verifier consumes and hashes all original bytes, checks a complete JPEG/PNG
container and PNG CRCs, checks dimensions against a conservative eight-byte per
pixel budget, and decodes the entire image. Filename and client MIME hints do
not supply trusted type facts. These are in-process resource bounds, not CPU or
process isolation. Format, size, source and busy failures use stable Blob codes;
unknown verifier errors are sanitized by the upload service.

The remaining eight extensions need production verifier adapters. Earlier PDF
fixture tests prove upload orchestration only. Lifecycle/GC and complete F2–F5
acceptance remain open.
