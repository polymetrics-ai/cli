# Verification checklist — issue-3752-rate-limit-admission-r1

Status: **scoped local verification complete; full CI/no-mistakes remains firstmate-gated**.

## Targeted behavior gates

- [x] Red B1 test ran before requester changes; its actual `30s` vs `90s` failure is retained in `TDD-LEDGER.md`.
- [x] Requester behavior suite passed via `go test ./internal/connectors/connsdk -count=1`.
- [x] Loader declaration suite passed via `go test ./internal/connectors/engine -count=1`.
- [x] Focused race suites passed for every new requester and loader test.
- [ ] Full `go test -race ./internal/connectors/connsdk -count=1` is not green: the unchanged
  `multipart_bounds_test.go:42,104` has an existing test-data race (the same lines are present at
  `origin/main`). This slice does not alter that test or multipart implementation; targeted new
  rate-limit tests pass under `-race` and the ordinary package suite passes.
- [x] Changed-package suites passed: `go test ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/bundleregistry ./cmd/connectorgen -count=1`.
- [x] `TestRequesterReturnsTypedRateLimitErrorAndObservation` proves a terminal 429 is both
  `*connsdk.RateLimitError` and its wrapped `*connsdk.HTTPError`, with no fixture credential text
  in `err.Error()` or typed observation values.
- [x] `TestRequesterAdmissionPreventsInitialLogicalSend` and
  `TestRequesterAdmissionHonorsCallerCancellationBeforeLogicalSend` prove a rejected/cancelled
  admission prevents the initial logical JSON, form, multipart, or stream request.
- [x] `TestRequesterAdmitsReplayableReadOncePerLogicalAttempt` proves a safe replayable read has
  one admission per logical `Client.Do` attempt even if `net/http` replays it; the ordinary and
  stream strict-write replay/redirect tests prove non-idempotent writes do not replay.

## Loader and fleet compatibility gates

- [x] `go test ./internal/connectors/engine -count=1`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — `550 connector(s) checked, 0 findings`.
- [x] `go run ./cmd/connectorgen surface-sync --check` — `550 connector(s) scanned`, no drift.
- [x] `rate_limits.json` is optional; no `internal/connectors/defs/<connector>/` declaration was
  changed. `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` prevents a future real
  declaration from being silently absent from production `defs.FS`.
- [x] Changed-path audit against `origin/main` found no `commandrunner`, command metadata,
  `api_surface`, `operations.json`, `cli_surface.json`, or connector-bundle migration path. The
  213-command runtime-preflight defect class is not expanded.

## Hygiene and project gates

- [x] `gofmt` ran on changed Go files.
- [x] Targeted `go vet` passed.
- [x] `go build ./cmd/pm` passed.
- [x] Individual non-full-suite gates passed: `tidy-check`, `lint`, `docs-check-no-build`,
  `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`,
  `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`.
- [x] `scripts/verify-gsd-workflow origin/main` passed with GSD/TDD evidence.
- [x] Changed-path audit confirms no connector migration, `commandrunner`, or deferred #3753/#3754/
  #3755 read/write/direct-operation/CLI activation edit.

## Rebase verification

- [x] Rebased the completed slice onto `origin/main` at `d215d9636` without a merge commit; the
  only conflict was the independently landed `Changefeed` bundle field, and both that field and
  this slice's `RateLimits` loader remain present.
- [x] Re-ran changed-package tests, focused `-race` requester/loader suites, targeted `go vet`,
  `go build ./cmd/pm`, all individual non-full-suite gates, and
  `scripts/verify-gsd-workflow origin/main` after the rebase.

## Deliberately not applicable in this slice

- CLI help/manual/website parity: #3755 owns the first operator-visible rate-limit surface. This
  foundation adds no command, flag, help topic, output format, or website documentation.
- Live/integration provider tests: prohibited. Only unit fixtures and `httptest` are permitted.
- Full `go test ./...` and monolithic `make verify`: CI owns these due the documented 550-connector
  timeout limitation; individual non-test gates remain local requirements.

## Final review sequence

The `execute-phase`, `verify-work`, and `code-review` prompts were generated and followed through
the documented inline fallback because Pi role spawning is forbidden by the delivery contract and
this task. `UAT.md` and `REVIEW.md` record the resulting automated evidence and review dispositions.
Do not start no-mistakes or a PR until firstmate directs the validation/ship stage.
