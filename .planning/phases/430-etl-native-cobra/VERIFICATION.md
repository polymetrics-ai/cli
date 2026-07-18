# Phase 430 Verification

Invocation session: `issue-430-pi-openai-codex-gpt-5.6-sol-high-20260718T225346Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `6c94754c58185df5aac53bd97587603c3154b1d5`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [x] Focused RED captured before production edits (`undefined: newETLCobraCommand`).
- [x] Native ETL/check/catalog/read/run/status/help tree; ETL legacy wrapper removed.
- [x] All current flags typed with repeated/bare/assigned compatibility.
- [x] Bare/text/JSON/long/short/positional help parity in focused tests.
- [x] Trailing help and literal `--` compatibility in focused tests.
- [x] Unknown flags, invalid actions, global assigned booleans, and no action-discovery bypass; test-first invalid-action trailing-help correction passes focused/repeated/race.
- [x] Batch-size parse/default/bounded flush behavior and configured sync validation in focused tests.
- [x] Cancellation, events, stdout/stderr, and one-envelope semantics in focused tests; broader telemetry gate pending.
- [x] Only ETL `parseFlags` call sites removed; dynamic connector parser remains.

## Gates

- [x] Focused native ETL/router tests (`13.396s`; broader ETL/router `27.999s`).
- [x] Focused repeated tests (`-count=5`: initial full contract `65.438s`; correction `1.061s`).
- [x] Focused race tests (initial full contract `146.473s`; correction `1.668s`).
- [x] Router and golden transcript tests; fixture unchanged (`21.327s`, correction preservation `6.749s`).
- [ ] Full `go test ./internal/cli/...`.
- [ ] Full app tests and ETL event/telemetry contracts.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify`.
- [ ] `git diff --check`; no dependency/unrelated/connector-def delta.

## CLI help/manual/website parity

- [ ] `pm help etl`.
- [ ] Bare `pm etl` exits 0 with contextual manual.
- [ ] `pm etl --help`, `-h`, positional `help`, and JSON manual routes.
- [ ] Invalid action remains usage error.
- [ ] `docs/cli/etl.md`: temporary generated diff clean or intentional update recorded.
- [ ] `website/content/docs/etl.mdx`: unchanged when no user-visible contract changed; generator/check recorded.
- [ ] Generated/golden help artifacts unchanged or explicitly reviewed.
- [ ] Completion discovery seam present; Phase 15 values remain deferred.
- [ ] Focused per-subcommand help/man churn remains deferred to Phase 19.

## Safety/scope/delivery

- [x] Exact branch/start and parent draft PR confirmed.
- [x] GSD doctor/list passed; unavailable programming-loop/manual fallback recorded.
- [x] Required Go CLI/testing/error/security/safety/context/concurrency/docs/Cobra skills loaded.
- [ ] Fixture/local temporary connectors only; no credentialed external checks.
- [ ] No optional services or service-backed runtime recording.
- [ ] No secret values requested, printed, summarized, stored, or logged.
- [ ] No dependencies, reverse execution, unrelated namespaces, or broad generated churn.
- [ ] Coherent planning, RED, GREEN, and final evidence commits pushed.
- [ ] No PR or review created.

Result: pending. `verificationPassed=false` until the complete declared gate set, including `make verify`, exits 0.
