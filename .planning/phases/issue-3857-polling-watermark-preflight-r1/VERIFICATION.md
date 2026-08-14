# Verification — #3857 polling-watermark preflight

## Required checks

| Area | Command/evidence | Status |
| --- | --- | --- |
| TDD red/green | Red is preserved in `TDD-LEDGER.md`; green: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestPollingPreflight|TestPollingModeEligibility|TestBundle.*Polling)'` | passed — exact happy/sad/edge fake assertions and all-mode real-preflight sweep |
| Immutable corpus | `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestPollingWatermarkConformanceSuiteRunsEveryMandatoryFixture$'` | passed — `0.736s`, all immutable v1 fixture IDs execute with no filter/skip surface |
| Database definition/schema | `go test -count=1 -timeout 20m ./internal/connectors/database` | passed — `8.305s` |
| Engine runtime | `go test -count=1 -timeout 20m ./internal/connectors/engine` | passed — `5.264s` |
| Connector and legacy adapter | `go test -count=1 -timeout 20m ./internal/connectors` | passed — `0.468s` |
| Commandrunner regression | `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'` | passed — `3.333s`; no REST preflight changes |
| App regression | `go test -count=1 -timeout 20m ./internal/app` | passed |
| Formatter and static analysis | `gofmt -l` on all changed Go files; `go vet ./...` | passed — no formatter output; vet exited 0 |
| Individual repository gates | `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check` | passed — lint reports `0 issues`; validation reports 552 bundles/0 findings; surface sync reports 0 drift |
| Build | `go build ./cmd/pm` | passed |
| CLI parity | `./pm help connectors`; `./pm connectors`; `./pm connectors --help`; `./pm connectors inspect postgres --json`; docs/website search | passed — existing metadata-only connector surfaces remain successful; no polling command or help output was introduced |
| GSD verification/review | generated `execute-phase`, `verify-work`, and `code-review` prompts executed by documented inline fallback; `REVIEW.md` | passed — no compatible isolated Pi runtime, no role spawning allowed; manual records are present |
| PR delivery | Open against `integration/4015-mvp-flat-r1`, read back `.base.ref` through `gh-axi api`, and wait for all checks green | pending — performed by the no-mistakes delivery pipeline after the committed handoff |

## Safety and scope checks

- No credentials, live database/provider calls, reverse ETL, target DML, or
  generic SQL/HTTP/shell pathway.
- `commandrunner` REST preflight is unchanged except only if a truthful native
  sync route is demonstrably required; no `api_surface` entry is added.
- No changes to #2986/#3745–#3749 changefeed capability derivation or #3762
  bounded-query taxonomy.
- No source/apply executor, driver, native-engine declaration, API surface, or
  polling CLI command was added. #3858/#3859 own execution and #3860 owns the
  public eligibility/help presentation and production runtime sweep.
