# REVIEW — issue #3867 rate-limit parking and automatic resumption

GSD code-review prompt: `scripts/gsd prompt code-review issue-3867-rate-limit-parking-r1`.
The canonical single-worker contract forbids role spawning here, so the standard review was completed inline after targeted race, vet, lint, build, and individual repository gates.

## Scope reviewed

- `internal/coordination/rate_parking.go`
- `internal/coordination/rate_parking_test.go`
- `internal/connectors/engine/rate_limit_parking.go`
- `internal/connectors/engine/rate_limit_parking_test.go`
- `.planning/phases/issue-3867-rate-limit-parking-r1/`

## Review method

1. Used CodeGraph to trace typed-error classification, durable store, scoped admission, timer, resume, and checkpoint call paths; inspected the final diff against `origin/integration/4015-mvp-flat-r1`.
2. Checked lock/store/scheduler ordering: a park store write happens before same-scope refusal, callbacks execute outside the coordinator lock, cancellation wins before state deletion/event emission, and successful resume deletes durable state before admission reopens.
3. Checked generic, unknown-source, reset-less, failed-callback, duplicate, cancellation, restart, and concurrent admission paths for zero unintended parking mutation/send behavior and secret disclosure.
4. Ran `git diff --check`, `go vet ./internal/coordination/... ./internal/connectors/engine/...`, targeted and repository lint, `go build ./cmd/pm`, the required package race command, and the individual non-suite repository gates listed in `VERIFICATION.md`.

## Findings and dispositions

| Severity | Finding | Disposition |
| --- | --- | --- |
| Info | Two newly added planning records used trailing Markdown whitespace, causing `git diff --check` to fail. | Fixed before review completion; `git diff --check` now passes. |

No unresolved security, correctness, concurrency, scope, or code-quality findings remain. The known `window_seconds` duration-overflow defect (#4125) was neither touched nor changed. No out-of-scope file requires a `needs-decision` record.
