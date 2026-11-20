
# TeaQL Tool API Reference

> [!WARNING]
> **DO NOT GUESS TOOL APIS**
> Do not guess generated or runtime tool methods.

> [!IMPORTANT]
> **USE TEAQL'S OWN TOOLING FIRST**
> Always use `cargo teaql` to retrieve assist content. Never bypass the CLI or invoke the generator endpoint directly.

To get the exact API usage and examples for TeaQL Tool APIs, execute the following command:

```bash
cargo teaql --input models/trace-chain-service.xml golang-assist-runtime-custom
```

> `models/trace-chain-service.xml` is the default model path. Adjust if your model file is located elsewhere.

Use the current Go Runtime Customization Assist for supported infrastructure.
Do not translate Rust tool methods into Go. If the required tool operation is
not documented, report `MISSING_ASSIST` instead of inspecting generated source.

Once the command succeeds, read its output. Use the printed code as a template to write your logic.

## Domain Object Assist APIs

For domain-object code, retrieve the current Go object/action Assist.

You can query these assist APIs for any object defined in your `models/trace-chain-service.xml`:

| Target | Description | Example Command |
|--------|-------------|-----------------|
| `golang-assist-query/[object]` | How to query and filter `[object]` | `cargo teaql --input models/trace-chain-service.xml golang-assist-query/school` |
| `golang-assist-create/[object]` | How to insert/create `[object]` | `cargo teaql --input models/trace-chain-service.xml golang-assist-create/school` |
| `golang-assist-update/[object]` | How to update `[object]` | `cargo teaql --input models/trace-chain-service.xml golang-assist-update/school` |
| `golang-assist-delete/[object]` | How to delete `[object]` | `cargo teaql --input models/trace-chain-service.xml golang-assist-delete/school` |
