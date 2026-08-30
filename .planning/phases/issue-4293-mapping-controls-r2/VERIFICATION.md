# Verification — Issue #4293 mapping controls R2

## Scope result

MAP-001/002/003 are repaired in mapping controls only.

- The canonical cohort manifest fixes the Batch R1 denominator at ten source
  locks and 4,341 source operations. It tracks each lock's raw-byte SHA-256,
  sorted source-ID SHA-256, and count, plus the aggregate sorted
  `connector<TAB>source_operation_id` SHA-256.
- The cohort references the canonical connector-local matrix input path for
  each connector, but it contains no matrix source-operation rows or lane
  cells. It therefore binds the composition convention without becoming a
  competing provider-fact denominator. Matrix presence and contents remain
  owned by the connector tracks.
- A source-lock-backed mutation needs explicit `direct_write` and
  `reverse_etl` dispositions. REST `PUT`/`PATCH`/`DELETE` and GraphQL
  `Mutation.*` roots are identity-backed mutations; POST remains explicitly
  source-node-classified so a non-mutating POST is not invented as a write.
- Record shape and ETL/binary/sync applicability are required, exact
  source-node-cited mapping facts. Candidate facts and dispositions that
  contradict collection pagination, binary media, or event/cursor evidence
  fail closed.

No connector definition, source lock, provider I/O, runtime engine execution,
credential, transport, certification, or generated connector output changed.
No runtime foundation is needed or proposed.

## Focused red → green proof

| Check | Result |
| --- | --- |
| `go test -timeout 20m ./cmd/connectorgen -run '^(TestBatch1SourceOperationMappingCohort|TestSourceOperationMapping)' -count=1` before implementation | Expected red: exit 1; missing `sourceOperationMappingCohortPathCheck`. |
| Same focused command after implementation | Green: exit 0 in 45.146s. Covers the deliberate missing/invalid cohort, mutation, citation, applicability, and traversal cases recorded in `TDD-LEDGER.md`. |
| `go test -timeout 20m ./internal/connectors/engine -count=1` | Green: exit 0 in 21.410s. |
| `go vet ./cmd/connectorgen ./internal/connectors/engine` | Green: exit 0. |
| `go run ./cmd/connectorgen source-operation-mapping-cohort data/connector-canon/batch1-source-operation-mapping-cohort.json --check` | Green: exit 0; `10 connector(s), 4341 source operation(s), 0 finding(s)`. |
| `go run ./cmd/connectorgen source-operation-mapping-cohort --help` | Green: exit 0; documents the check-only form. |
| `go run ./cmd/connectorgen --help` | Green: exit 0; lists `source-operation-mapping-cohort`. |
| `go run ./cmd/connectorgen declaration-admission` | Green: exit 0; `1 connector(s), 1 source operation(s), 0 finding(s)`. This is the existing scoped declaration-admission projection, not the Batch R1 denominator check. |
| `go run ./cmd/agentcontractgen check` | Green: exit 0; canonical contract and registered projections are current. |
| `jq empty` for both mapping schemas and the cohort JSON; `gofmt`; `git diff --check` | Green: exit 0. |

## Self-review

- Denominator: the cohort checker rejects altered exact membership, per-lock
  digest/count/source-ID digest, aggregate count/digest, and canonical path
  traversal. It reads current source locks only and never parses/copies a
  connector-local matrix.
- Mutation control: fixture tests reject both missing lanes for REST mutation
  and GraphQL mutation roots; a locked REST delete test prevents relying on a
  POST-only heuristic. The cited non-mutating POST test remains green.
- Source-node control: every mapping fact and `not_applicable` reason citation
  must match both the locked source URL and its exact source-lock location.
  Focused negatives cover missing record shape, wrong-node record-shape and
  ETL citations, binary download/upload, and sync contradictions.
- Scope: the diff is limited to mapping control command/schema/test code, the
  cohort artifact, and planning evidence. It contains no imports or calls to
  runtime execution, credentials, transport, provider I/O, certification, or
  connector definition generation.

## Broader baseline

No broad suite was run for this scoped change. The frozen R1 review recorded
separate pre-existing broad `cmd/connectorgen` parity failures (six parity
cases), operation-evidence drift, and connector-definition findings. Those
are not modified or reclassified here and are outside the mapping-control
scope.
