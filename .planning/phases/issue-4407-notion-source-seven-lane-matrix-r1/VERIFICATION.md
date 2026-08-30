# Verification

## Green checks

- `jq empty internal/connectors/defs/notion/sources/notion-operation-source-lock.json internal/connectors/defs/notion/sources/notion-operation-crosswalk.json internal/connectors/defs/notion/sources/notion-source-lane-matrix.json` — passed.
- `go test -timeout 20m ./internal/connectors/defs/notion` — passed.
- `go vet ./internal/connectors/defs/notion` — passed.
- `go run ./cmd/agentcontractgen check --root .` — passed: canonical contract and registered projections are current.
- `git diff --check` — passed.

## Source reconciliation

- Retained lock denominator: 49 rows; matrix: 49 rows × 7 lanes = 343 cells.
- Crosswalk: 49 exact source rows, 2 source-bound surface-only identities, and 0 lock-only rows.
- Matrix lane dispositions: direct read 20 mapped-unproven / 29 not-applicable; direct write 25 / 24; binary download 0 / 49; binary upload 1 / 48; ETL 12 / 37; reverse ETL 25 / 24; sync transport 0 / 49.
- The test includes in-memory red checks for a hidden source row, missing lane, missing backlink, unsupported `implemented` promotion, dropped crosswalk boundary, and dropped mapping-control restriction.

## Check-only restrictions (not runtime gaps)

- `go run ./cmd/connectorgen validate internal/connectors/defs/notion --json` reports `parse source lock: json: unknown field "source_operation"` for the retained v2 lock.
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` reports `notion source projection: canonical source descriptor is missing`.

Both are explicitly recorded as all-row `mapping_restriction` records in the matrix. No shared importer, executor, runtime, certification, or CLI behavior was changed.
