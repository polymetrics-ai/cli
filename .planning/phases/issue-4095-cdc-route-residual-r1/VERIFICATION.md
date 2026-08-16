# Verification — Issue 4095 residual route

| Acceptance criterion | Result | Evidence |
| --- | --- | --- |
| Live PostgreSQL CDC → PostgreSQL R4 route | not_run | Pending a tagged PostgreSQL 14+ pgoutput test with independent target read-back. |
| Receipt precedes LSN acknowledgement; restart and replay are covered | not_run | Pending the same whole-route test; component tests alone are insufficient. |
| Enumerated R1/R2/destination `change_capture` pre-I/O matrix | not_run | Pending named rows with typed error and source/target I/O counters. |
| R3 PostgreSQL CDC → API | non_executable | GitHub destination `sync_transport.json` has no CDC source binding and declares deletes unavailable; `change_capture` is source-only to the connection-owned local warehouse. This task intentionally adds no API writer/action. |
| No surface overstates R3 support | not_run | Pending declaration/capability surface inspection after implementation. |

## Required gates

- Targeted Go test commands for every changed package and `./internal/cli`.
- `go test -timeout 20m ./cmd/connectorgen` (consumer-package rule).
- Tagged PostgreSQL live test with the shared Docker endpoint; any unavailable endpoint will be recorded as `not_run` with the exact reason.
- `gofmt`, `go vet`, `go build ./cmd/pm`, individual `make verify` non-test gates, generator byte-stability checks, connector boundary, and website docs generation twice if repository policy requires it.

## Execution record

No acceptance gate has been recorded as passing before it is executed.
