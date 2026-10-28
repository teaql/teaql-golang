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

`NewMediaContentVerifier` is a separate Linux-only MP3/MP4 adapter. The host
supplies absolute paths to trusted `ffprobe`, `ffmpeg` and `prlimit` binaries,
keeps them patched and passes input, frame, pixel and concurrency limits. It
walks complete MP3 frames or MP4 boxes, probes actual streams and frame counts,
requires a non-cover-art MP4 video track, then decodes all media from a pipe.
No client filename or URL reaches FFmpeg. `prlimit` caps each child at 10 CPU
seconds and 512 MiB virtual memory; a parent timeout limits wall time to 15
seconds. The verifier hashes original bytes and reports `audio/mpeg` or
`video/mp4`. This is format validation, not malware scanning. Its readiness
check is called by `BlobReferenceModule.Ready` when that verifier is installed.

`NewOfficeContentVerifier` is a separate standard-library DOCX/XLSX parser. It
requires a complete ZIP central directory and fully reads every member with
CRC checks under compressed input, expanded-byte, part, XML-element, depth and
concurrency limits. It rejects duplicate or traversal part names, XML
directives/entities, malformed XML, mismatched Office content types and main
relationships; DOCX needs a Word document body, while XLSX needs a workbook
whose sheet IDs resolve to real worksheet roots. It hashes the original package
and reports the verified DOCX/XLSX MIME. This checks OOXML package structure,
not macros, malware or every application-level Office constraint.

`NewPDFContentVerifier` is a separate Linux-only PDF adapter. The host supplies
absolute paths to patched `pdfinfo`, `pdftoppm` and `prlimit` binaries, plus
input, page, rendered-output and concurrency limits. It checks the PDF envelope,
requires an unencrypted page count within policy, renders every page through
stdin to bounded 64-pixel PNG output and decodes every resulting image. Each
child has CPU and address-space limits and the verification has a wall timeout;
document bytes and rendered pages are not written to scratch. It hashes the
original PDF and reports `application/pdf`. This validates decodability within
the configured limits; it is not malware scanning or a universal PDF conformance
check. The host must call readiness before accepting uploads.

To accept several formats through one upload service, the host
constructs `NewCompositeContentVerifier` with separate image, media, Office and PDF
`ContentVerifierRoute` groups, then passes the composite to
`NewSQLiteUploadService`. Duplicate, invalid and nil routes fail at assembly;
unregistered extensions fail before reading the source. Its explicit readiness
check forwards to each configured parser. This router does not load optional
parsers or expand the set of implemented formats.

The remaining two extensions (DOC/XLS) need production verifier adapters.
Earlier PDF fixture tests prove upload orchestration only.
Lifecycle/GC and complete F2–F5 acceptance remain open.

`NewSQLiteFileGenerationSweepService` is an optional, host-scheduled File
lifecycle operation. The host must first upgrade every process sharing the
private File root to the writer-lock protocol and explicitly assert that fact.
`RunBatch` requires the actor's exact `FileObjectStorage`, current maintenance
authorization, a trusted scope, and no active graph SQL transaction. It pages
server-stream catalog rows, including terminal deleted rows, then calls the
provider's fenced per-key old-generation sweep. A failed key is reported with
the stable `BLOB_STORE_FAILED` code; the host can retry from the beginning.
The current data generation stays protected by the provider's writer and
manifest locks. This service does not collect direct-upload objects or start a
background worker.
