# Code review — Issue #4176

## Mode

Inline/manual `code-review` fallback. The generated GSD prompt was resolved before implementation; the canonical delivery contract and this runtime require one active worker, so no separate reviewer worker was available.

## Scope reviewed

- Stage ordering, error aggregation, and cleanup in `stageScheduleRoundtrip`.
- Report semantics for direct fire versus unstarted scheduler daemon.
- Scripted stage contract and real CLI construction-path coverage.
- Happy, bad, and edge test assertions; report persistence implications; secret-safe output.

## Findings

No critical, warning, or informational finding remains.

- The full-stage omission from the audit does not reproduce on the dispatch base: `runFullReadSweep` already invokes the glue stages per catalog stream. The new report comparison guards both the tail and full-sweep execution paths.
- A failed install assertion now prevents the `schedule fire` command from being invoked. The test verifies zero fire calls, aggregate failure, and cleanup, so removal cannot hide an unavailable execution path.
- The direct-fire stage checks a real product CLI envelope for both flow and durable fire status. It does not claim that a crontab/systemd/Temporal daemon triggered the entry; the persisted capability result is `not_live`.
- No credential values, approval material, raw provider data, or scheduler configuration were introduced. The existing authority redaction assertion also covers the new fire envelope.

## Review evidence

- `go test -timeout 20m ./internal/connectors/certify -count=1` — pass.
- `go test -timeout 20m ./internal/cli -count=1` — pass.
- `go vet ./internal/connectors/certify ./internal/cli`, `go build ./cmd/pm`, `make lint`, and `git diff --check` — pass.
