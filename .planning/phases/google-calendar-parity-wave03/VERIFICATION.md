# Google Calendar parity resume — verification checklist

## Completed focused gates before the write-action slice

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 550 connectors, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync` and `--check` — 550 connectors, 0 updates/drift.
- [x] `go test ./internal/connectors/conformance/...` — pass after restoring the recovered fixture-auth sentinel.
- [x] `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` — pass.
- [x] `go test ./internal/connectors/hooks/google-calendar ./internal/connectors/bundleregistry ./internal/connectors/native/nativeset` — pass.
- [x] `go test ./internal/cli/...` — pass, including regenerated golden transcripts.
- [x] `go vet` for changed packages and `go build ./cmd/pm` — pass.
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`, `make connector-boundary`, and `make release-workflow-check` — pass.
- [x] `cd website && pnpm run gen:website-data && pnpm run typecheck` — pass; generated data committed.
- [x] `pm help google-calendar`, bare `pm google-calendar`, and `pm google-calendar freebusy query --help` — pass; the required flags render as required. A no-flag FreeBusy request and a removed mutation command both fail without any credential/provider access.
- [x] `git diff --check` — pass.

## Current ledger evidence

- Official Discovery total: 38 operations (11 GET, 15 POST, 4 PATCH, 4 PUT, 4 DELETE).
- Executable now: 38 (11 streams, `freeBusy.query`, and 26 reverse-ETL write actions).
- Blocked: 0. The 2026-08-05 correction required all 26 record-shaped mutations to use the current `writes.json` executor; none requires `rest_write`.
- Planned: 0.
- Provider citations: 149/149 declared request-field uses (27 read/direct-read + 122 writes) and 38/38 operation-level sources, recorded in `REQUEST-FIELD-RESEARCH.md` and `google-calendar-official-operations.json`.

## Write-action slice gates

- [x] Red: `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar` rejected a `writes.json` action without matching API-surface coverage.
- [x] Red: fixture-backed conformance rejected unsupported `minLength`/`const` schema keywords; schemas now use the supported `enum` form for channel type.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar` — 1 connector, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync internal/connectors/defs/google-calendar --check` — 0 drift.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/google-calendar' -count=1` — 26 write fixtures and the read surface pass.
- [x] `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1` — every implemented command, including write actions, passes real preflight.
- [x] `go build ./cmd/pm` — pass.
- [x] `go test ./internal/connectors/hooks/google-calendar ./internal/connectors/bundleregistry ./internal/connectors/native/nativeset -count=1` — pass, including the 26-action inventory and three query-on-wire regressions.
- [x] `go test ./internal/cli/... -count=1` — all non-golden cases passed; the only failures were stale root help fixtures that correctly lacked the new Google Calendar write summary. Regenerated with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1`, then `TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals` passed.
- [x] `go build ./cmd/pm`, `pm google-calendar --help`, `pm google-calendar events insert --help`, and `pm google-calendar calendars transfer-ownership --help` — pass and render one non-duplicated command group per resource plus typed action flags and confirmations.
- [x] `pm connectors inspect google-calendar --json` — reports `capabilities.write=true` and all 26 `write_actions`.
- [x] `go vet ./internal/connectors/conformance ./internal/connectors/commandrunner ./internal/connectors/hooks/google-calendar ./internal/cli` — pass.
- [x] `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`, `./pm docs validate --connectors-dir docs/connectors`, and `make docs-check-no-build` — pass; generated Google Calendar manual, skill, and catalogs are tracked.
- [x] `cd website && pnpm run gen:website-data && pnpm run typecheck` — pass; the generated catalog records Google Calendar as write-capable.
- [x] `make tidy-check`, `make lint`, `make smoke-no-build`, `make connector-boundary`, and `make release-workflow-check` — pass.
- [x] `git diff --check` — pass.

## Review remediation gate

- [x] `go test ./internal/connectors/hooks/google-calendar -count=1` — pass; metadata, fresh event reads, legacy projections, and settings pagination are covered.

## CI batchability remediation

- [x] Red: `go test ./internal/connectors/engine -run '^TestEveryShippedWriteActionIsBatchable$' -count=1` reproduced all ten intentional Google Calendar `batchable:false` failures.
- [x] `go test ./internal/connectors/engine -run '^TestEveryShippedWriteActionHasExpectedBatchability$' -count=1` — pass.
- [x] `go test ./internal/connectors/engine -count=1` — pass.
- [x] `go test ./internal/app -run '^(TestPlanReverseETLRefusesNonBatchableAction|TestRunReverseETLRefusesStoredNonBatchablePlan|TestNonBatchableActionStillExecutesAsConnectorCommand)$' -count=1` — pass.
- [x] `go test ./internal/connectors/hooks/google-calendar -count=1` — pass.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/google-calendar' -count=1` — pass.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar` — 1 connector, 0 findings.
- [x] `make connector-boundary` — clean (550 connectors, 0 findings/warnings).
- [x] `go vet ./internal/connectors/engine`, `gofmt -l internal/connectors/engine/batchable_test.go`, and `git diff --check` — pass/clean.
- [ ] Full `make verify` rerun — intentionally left to the outer no-mistakes CI phase owner; focused local evidence is not recorded as full verification.

## Constraints

No credentialed provider checks and no write execution were run. A representative write preview intentionally stopped before any provider interaction because this disposable worktree has no `.polymetrics` project. Do not claim full `make verify` or `go test ./...`; the shared parity contract requires focused gates because the full suite exceeds bounded command windows, and the outer no-mistakes executor owns the next full CI run.
