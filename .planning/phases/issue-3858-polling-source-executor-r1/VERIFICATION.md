# Verification — #3858 page-safe polling source executor

## Required checks

| Area | Command/evidence | Status |
| --- | --- | --- |
| TDD red/green | `TDD-LEDGER.md` with observable source, sink, and checkpoint assertions | pass — first focused run failed to compile against the intentionally absent executor; the real-executor focused suite is green. |
| Source executor | `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestPollingSourceExecutor'` | pass |
| Engine and connector regressions | `go test -count=1 -timeout 20m ./internal/connectors/engine/... ./internal/connectors/...` | pass |
| App regression | `go test -count=1 -timeout 20m ./internal/app` | pass |
| Static analysis | `gofmt -w internal/connectors/engine/polling_source.go internal/connectors/engine/polling_source_test.go`; `go vet ./...` | pass |
| Individual repository gates | `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check` | pass |
| Build | `go build ./cmd/pm` | pass |
| GSD verification/review | Inline manual fallback plus `REVIEW.md` | pass — manual standard review found no remaining actionable findings. |
| PR delivery | Push task branch, create the PR against the integration base, and API read-back `.base.ref` | pending |

## Safety and scope checks

- No credentials, live database/provider calls, native driver changes,
  reverse ETL, target DML, raw SQL, generic HTTP, shell, or path surface.
- No implementation change outside this shared source executor and its tests.
- The legacy #3880 changefeed executor is not used as a new public polling or
  CDC claim; #3860 retains public-surface ownership.
- The issue explicitly excludes a live database/native adapter. The executor's
  real side effects are proved with a structured fake runner and the shared
  #3810 durable acknowledgement helper: page requests, destination sends, and
  checkpoint-store mutations are independently observable without credentials
  or a provider connection.
