# TDD ledger — Zoom full definition mapping

Issue: #4265
Phase: `zoom-full-definition-mapping-r1`

## Source-lock integrity

Red: the first attempted source lock serialized zero modules, so it could not substantiate any
declaration.

Green: the corrected public-only acquisition pins 35 module documents, their individual and
aggregate SHA-256 digests, byte counts, OpenAPI versions, server roots, method counts, and 1,937
source operations. An exact method/origin-path comparison finds 1,911 of 1,913 ledger identities:
1,908 blocked rows plus the three preserved ETL rows. The remaining two ledger-only Phone paths and
26 source-only paths are committed as explicit reconciliation records.

Refactor: server-root normalization is recorded in the source lock (`server_url`, `path`, and
`source_path`) so the standard `/v2` routes and the key-connector `/api/v2` route are compared
without changing their provider path contracts.

## Declaration contract tests

Red: `TestSourceBackedOperationInventoryKeepsEveryContractVisible` first observed zero source-backed
operation contracts (wanted 1,820 candidates after protected/external cohorts), and the reverse-ETL
fixture test observed zero warehouse destination actions (wanted two). A first generated operation
pass also failed the operations schema because provider parameter enum values included non-strings;
the source fact is retained in the crosswalk, while the unsupported field is omitted rather than
coerced in `operations.json`.

Green: the connector-local inventory now contains 1,748 source-backed operation contracts (776
`rest_read`, 971 `rest_write`, and one `binary_download`) with 311 typed destructive DELETE
contracts. The two real destination actions run through plan/preview and a loopback HTTP fixture.
`TestPinnedSourceCrosswalkAccountsForEveryIdentity` proves the 35-document, 1,937-operation lock
and every 1,911 exact / 26 source-only / two ledger-only identity result.
`TestDeclarationDispositionAccountsForThePinnedSourceAndLedger` proves the 1,913 ledger rows plus
26 source-only rows have a declared or disabled result, and every disabled result has a
fixed-vocabulary reason, evidence, and recovery state. `go test -timeout 20m
./internal/connectors/defs/zoom -count=1`, `go run ./cmd/connectorgen validate
internal/connectors/defs/zoom --json`, and `go run ./cmd/connectorgen surface-sync
internal/connectors/defs --check` pass.

The intentionally absent `sync_transport.json` is not a missing declaration: a closed
`declarative_stream_source` executor and a durable source-to-warehouse conformance proof do not
exist. `sources/zoom-foundation-gaps.json` records this along with the REST-write, upload, and
binary redirect/origin gaps. Zoom's later central scope admission requires only the generated local
certification matrix; no authentication, engine, generator, allowlist, or status file is changed by
this lane.

Refactor criteria: run `connectorgen validate`, `connectorgen surface-sync --check`,
`make connector-boundary`, and then the full `make verify` before a push.
