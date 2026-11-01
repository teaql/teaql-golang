# Generated graph save and loaded-view regression

```sh
node examples/fix-e-alignment/verify-local.cjs
```

This builds the retained evaluated fixture against the local checkout, executes
the application twice without resetting its SQLite stores, and removes its owned
compiler outputs after the processes exit. Compact evidence and databases remain.

`partial_root_test.go` loads two untouched sparse ancestors and one complete
modified leaf through generated Q/E APIs. A second real root is created through
the generated audited-save API. Both wrong-context attempts must reject before
all ten observed provider/allocation/policy/audit boundaries, preserve pending
intent and the original JSON/E view, and retain the complete selected error path.
The same graph then retries successfully with its correct context; only the
child's stored value and version change.

The counted test ID SPI deliberately reserves disjoint ranges for each round;
it tests ownership admission, not persistent or early factory ID allocation.
Independent read-only SQL is an oracle only; setup and mutations use generated
APIs. No generated implementation is hand-edited.
