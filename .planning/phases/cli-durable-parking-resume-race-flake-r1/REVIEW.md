# Code Review — durable parking resume claim race

## Scope

- `internal/app/app.go`
- `internal/app/state_lock_test.go`
- `internal/cli/durable_parking_cli_process_test.go`
- phase evidence

## Inline review result

**Clean.** The correction does not add a retry, sleep, timeout, skip, or
quarantine. `app.Open` reads only the atomically replaced snapshot while
normalization still persists through the existing mutation lock. The direct
process edge test preserves two concurrent helpers and the exact one-send
assertion. The malformed-state refusal remains before durable coordination
side effects. No dependency, credential, public CLI, documentation, or
generated-artifact change was introduced.

## Automated/static evidence

- `go vet ./internal/app ./internal/cli` — pass.
- `make lint` — pass, 0 issues.
- `git diff --check` — pass.
- Targeted and full changed-package tests — pass; process race certification
  passed 20/20 under recorded load.

## GSD code-review route

`scripts/gsd prompt code-review cli-durable-parking-resume-race-flake-r1` was
resolved. The canonical inline/manual fallback applies because role spawning is
not available or permitted for this single-worker delivery. No findings remain.
