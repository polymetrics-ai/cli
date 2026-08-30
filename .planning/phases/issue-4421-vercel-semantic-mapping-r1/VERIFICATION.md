# Verification — Issue #4421 Vercel semantic mapping repair

## Scoped results

| Check | Result |
| --- | --- |
| `go test -count=1 -run '^(TestVercelSourceLaneMatrixRetainsEveryLockedOperationAndLane|TestVercelSourceLaneMatrixUsesRetainedSourceSemanticsBeyondHTTPVerb|TestVercelSourceSemanticLaneEvidenceRejectsIncompleteOrContradictoryFacts)$' ./internal/connectors/defs/vercel` | PASS |
| `go test -count=1 ./internal/connectors/defs/vercel` | PASS |
| `go test -race -count=1 ./internal/connectors/defs/vercel` | PASS |
| `go vet ./internal/connectors/defs/vercel` | PASS |
| `jq empty internal/connectors/defs/vercel/sources/vercel-source-lane-matrix.json` | PASS |
| `go run ./cmd/agentcontractgen check` | PASS — canonical contract and registered projections are current |
| `git diff --check` | PASS |
| source-lock and crosswalk diff | PASS — zero changed bytes |

The matrix validator recomputes the 400 retained source rows, all 2,800 lane cells, every mapping backlink, and the summary counters. The semantic tests cover happy, bad, and edge evidence as recorded in `TDD-LEDGER.md`.

## Full repository suite

`go test -count=1 ./...` was run to completion and returned non-zero. The Vercel definition package passed in that same output. Reported failures are outside this repair’s changed files:

- `cmd/connectorgen`: GitLab fleet-count expectations (`249` versus `246` aliases; `967` versus `733` source-backed rows) and a ten-minute timeout in `TestOperationEvidencePreservesSavedAndInteractiveLanes`.
- `internal/connectors/defs`: unclassified `asana/enabled_connector_contract.json` in the production embed inventory.
- `internal/synctransport`: `TestTransport_ReadBackGetsIndependentUnitDeadline/full-overwrite` received `context deadline exceeded`.

Those failures are not repaired or masked here. Since no baseline replay of the global suite was performed, the accurate status is **scoped green; repository-wide suite red for out-of-scope failures**.

## Residual restrictions

- These cells are `mapped_unproven`, not executable runtime claims.
- `application/gzip` remains preserved as provider source evidence; shared MIME admission is explicitly out of scope and must be handled by its separate repair.
- `go run ./cmd/connectorgen validate internal/connectors/defs/vercel --json` remains red because the frozen Vercel definition lacks `sources/vercel-operation-descriptor.json` (`source_projection: canonical source descriptor is missing`). Adding that descriptor is outside this connector-local matrix repair.
- No ETL or sync-transport status was promoted without retained pagination or webhook/runtime facts.
