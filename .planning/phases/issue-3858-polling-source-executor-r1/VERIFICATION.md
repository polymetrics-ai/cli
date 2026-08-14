# Verification — #3858 page-safe polling source executor

## Required checks

| Area | Command/evidence | Status |
| --- | --- | --- |
| TDD red/green | `TDD-LEDGER.md` with observable source, sink, and checkpoint assertions | pending |
| Source executor | `go test -count=1 -timeout 20m ./internal/connectors/engine -run 'Polling.*Source|PollingWatermarkConformance'` | pending |
| Engine and connector regressions | `go test -count=1 -timeout 20m ./internal/connectors/engine/... ./internal/connectors/...` | pending |
| App regression | `go test -count=1 -timeout 20m ./internal/app` | pending |
| Static analysis | `gofmt -l <changed Go files>` and `go vet ./...` | pending |
| Individual repository gates | `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check` | pending |
| Build | `go build ./cmd/pm` | pending |
| GSD verification/review | Inline manual fallback plus `REVIEW.md` | pending |
| PR delivery | Push task branch, create the PR against the integration base, and API read-back `.base.ref` | pending |

## Safety and scope checks

- No credentials, live database/provider calls, native driver changes,
  reverse ETL, target DML, raw SQL, generic HTTP, shell, or path surface.
- No implementation change outside this shared source executor and its tests.
- The legacy #3880 changefeed executor is not used as a new public polling or
  CDC claim; #3860 retains public-surface ownership.
