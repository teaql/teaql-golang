# Natural Checker message example

Use this repository's local runtime via `verify-local.cjs`. `fixture/` is a
read-only generated library, not an application customization surface. Fix its
upstream model/generator and regenerate if needed; keep manifests and provenance.

Application assertions belong in `natural_messages_test.go`. Do not replace the
generated Checkers, install presentation metadata or rewrite sensitivity and
constraints to make this automatic-composition example pass. Use bounded Q,
loaded E and audited Save APIs; preserve the generated diagnostic type.
