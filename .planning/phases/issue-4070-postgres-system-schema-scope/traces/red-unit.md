# RED — pre-connection system-schema rejection

**Candidate:** `49a9386d2c629e53594c6bba1dd9a74a05b3bff5` plus the committed-RED test-only work.
**Production code modified:** no.

## Command

```sh
go test -count=1 \
  -run '^TestTypedCatalogRejectsReservedConfiguredSchemasBeforeConnect$' \
  ./internal/connectors/native/postgres
```

## Observed result

Exit status: `1` (expected RED).

Every required case — `pg_catalog`, `information_schema`, `pg_toast`,
`pg_toast_4070`, and `pg_temp_4070` — failed because the candidate entered the
typed-catalog snapshot and attempted a loopback TCP connection. The safe error
shape was:

```text
catalog postgres: begin typed catalog snapshot: ... dial tcp 127.0.0.1:<run-port>: connection refused
```

That is the wrong boundary: a reserved schema has reached pool/transaction
creation instead of returning the desired identifier-free scope error. The test
is deliberately behavioral: it expects the scope message rather than referring
to a future production symbol, so it proves the current path actually reaches
transport.
