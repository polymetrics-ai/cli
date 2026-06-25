# TDD Ledger

Phase: native-go-all-connectors

## Red Evidence

- Changed catalog tests to require the true embedded status split: 1 enabled connector and 646 planned native ports.
- Changed registry tests to require only enabled catalog connectors to be executable.
- Changed fixture conformance tests to prove fixture scaffolds are metadata/docs/redaction coverage, not runtime enablement.
- Changed CLI tests to reject direct ETL for `source-stripe` while it remains `planned_native_port`.
- Changed reverse ETL tests to reject credentials and writes for `destination-postgres` while it remains `planned_native_port`.

The first focused CLI run exposed the stale reverse ETL assumption:

```text
--- FAIL: TestReverseETLToNativeCatalogDestinationWritesReceiptAfterApproval
connector "destination-postgres" not found
```

That failure was expected after removing planned connectors from the executable registry.

## Green Evidence

Connector and CLI tests passed after adding the registry gate, preserving fixture conformance, and adding the `source-github` alias:

```text
go test ./internal/connectors
ok polymetrics/internal/connectors

go test ./internal/cli
ok polymetrics/internal/cli
```

Full verification passed:

```text
make verify
go test ./...
Validated connector docs in docs/connectors
smoke ok
```
