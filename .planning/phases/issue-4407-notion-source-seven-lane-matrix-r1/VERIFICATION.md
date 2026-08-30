# Verification

## Green checks

- `jq empty internal/connectors/defs/notion/sources/notion-operation-source-lock.json internal/connectors/defs/notion/sources/notion-operation-crosswalk.json internal/connectors/defs/notion/sources/notion-source-lane-matrix.json` — passed.
- `go test -timeout 20m ./internal/connectors/defs/notion -run TestNotionSourceLaneMatrixRetainsEveryLockedOperationAndLane -count=1` — red before matrix repair (`post-database-query` was not mapped), then passed after the semantic POST direct-read cells were repaired.
- `go test -timeout 20m ./internal/connectors/defs/notion` — passed.
- `go test -race -timeout 20m ./internal/connectors/defs/notion` — passed.
- `go vet ./internal/connectors/defs/notion` — passed.
- `go test -timeout 20m ./cmd/connectorgen -run 'TestNotionAPISurface|TestNotion' -count=1` — passed.
- `go run ./cmd/agentcontractgen check --root .` — passed: canonical contract and registered projections are current.
- `git diff --check` — passed.

## Source reconciliation

- Retained lock denominator: 49 rows; matrix: 49 rows × 7 lanes = 343 cells.
- Crosswalk: 49 exact source rows, 2 source-bound surface-only identities, and 0 lock-only rows.
- Matrix lane dispositions: direct read 24 mapped-unproven / 25 not-applicable; direct write 25 / 24; binary download 0 / 49; binary upload 1 / 48; ETL 12 / 37; reverse ETL 25 / 24; sync transport 0 / 49.
- The direct-read candidates include four source-documented semantic POST reads: `post-database-query`, `post-search`, `query-meeting-notes`, and `introspect-token`. Each remains `mapped_unproven`; this is mapping evidence, not a runtime availability claim.
- `query-meeting-notes` is both a bounded direct-read candidate and a mapped-unproven ETL candidate. Its ETL cell retains the source limitation that `has_more` has no retained continuation request input.
- The test includes in-memory red checks for a hidden source row, missing lane, missing backlink, unsupported `implemented` promotion, dropped crosswalk boundary, dropped mapping-control restriction, semantic POST direct-read acceptance, mutation-POST rejection, and a removed semantic POST response fact.

## Check-only restrictions (not runtime gaps)

- `go run ./cmd/connectorgen validate internal/connectors/defs/notion --json` reports `parse source lock: json: unknown field "source_operation"` for the retained v2 lock.
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` reports `notion source projection: canonical source descriptor is missing`.

Both are explicitly recorded as all-row `mapping_restriction` records in the matrix. No shared importer, executor, runtime, certification, or CLI behavior was changed.

## CLI help/manual/website parity

Not applicable. This repair changes only a source-lane mapping artifact and its connector-local test; it does not add or alter a `pm` command, flag, generated manual, or website documentation surface.
