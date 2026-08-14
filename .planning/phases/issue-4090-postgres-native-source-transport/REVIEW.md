# Inline deep review — Issue #4090

## Scope

- `internal/connectors/native/postgres/connector.go`
- `internal/connectors/native/postgres/transport_source.go`
- `internal/connectors/native/postgres/transport_source_test.go`
- `internal/connectors/native/postgres/dynamic_catalog_integration_test.go`

This is a manual inline review because the canonical contract forbids spawning
review roles for this single-worker issue phase. The review covered descriptor
selection, preflight ordering, transaction/cancellation behavior, SQL-identifier
safety, catalog/order invariants, page bounds, checkpoint structure, secret
handling, and the live test’s registry path.

## Findings and dispositions

| Severity | Finding | Disposition |
| --- | --- | --- |
| Warning | Checkpoint dedupe JSON-encoded stable-key values, which rejects a valid PostgreSQL `float8 NaN` key. | Fixed. Dedupe is now derived from the repeatable-read barrier, typed schema, source identity, and page ordinal; `TestPostgresSnapshotCheckpointDoesNotRequireJSONEncodableKeyValues` is green. |
| Warning | The checkpoint source scope treated the catalog component as a SQL identifier and rejected dynamic-catalog database names such as `analytics-db`. | Fixed. Only schema/relation are SQL identifiers; the catalog component is identity-only and is tested by `TestPostgresSnapshotRelationRefAllowsTypedCatalogDatabaseName`. |

No critical, high, or remaining warning findings are open.

## Security and boundary result

- All SQL identifiers rendered by the executor originate from the already typed
  catalog and retain strict PostgreSQL identifier validation. Query values and
  page limits are pgx parameters; no raw caller SQL exists.
- Source identity is validated before pool open. Missing descriptor, wrong
  family, and unregistered executor are separately proven at registry preflight
  with a source that records every I/O attempt.
- No secret/config value is logged. The live trace contains fixture row IDs and
  opaque checkpoint metadata only.
- `internal/connectors/engine/bundle.go` is untouched; no App composition,
  generic polling, target, CDC, certification-version, or generic transport
  change was introduced.

## Review validation

`go test`, `go test -race`, `go vet`, `make lint`, all individual required
repository gates, and the final live PostgreSQL 16.10 dbtest command in
`VERIFICATION.md` passed after these fixes.
