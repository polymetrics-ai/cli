# Phase 437 Verification

Invocation `issue-437-pi-sol-high-20260719T095145Z`; Sol/high; start `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`.

`verificationPassed`: false

## TDD / behavior

- [x] Six artifacts predate tests and production edits.
- [ ] RED captured before production edit.
- [ ] Native connectors actions and nested certify actions/current flags/operands.
- [ ] Bare/text/JSON/topic/positional/trailing help; invalid action usage.
- [ ] Literal `--`, malformed/legal unknown, action/operand discovery and globals.
- [ ] Certify exits 0/1/2/3 and one-envelope semantics.
- [ ] Re-entrancy, bounded concurrency, cancellation, events, telemetry.
- [ ] Credential values absent from all output/artifacts/signals.
- [ ] Only namespace parser calls removed; dynamic connector parser unchanged.

## Focused / broad gates

- [ ] Focused connectors/certify native tests.
- [ ] Focused repeated `-count=10`.
- [ ] Focused `-race`.
- [ ] Router and golden tests.
- [ ] Full `go test ./internal/cli/...` and `go test ./internal/connectors/certify`.
- [ ] Required certify smoke.
- [ ] `go run ./cmd/connectorgen validate` / connector validation.
- [ ] `gofmt -w cmd internal`; clean formatting diff.
- [ ] `go vet ./...`.
- [ ] `go test ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify` exit 0.

## Help/manual/website parity

- [ ] `pm help connectors`.
- [ ] bare `pm connectors`.
- [ ] `pm connectors --help` and positional/trailing help.
- [ ] certify command/help examples and exits accurate.
- [ ] `docs/cli/connectors.md` regenerated.
- [ ] website CLI reference mirrored/generated with no unrelated churn.
- [ ] golden/manual generation checks pass.
- [ ] Completion/discovery remains current; no Phase 15/19 churn.

## Safety/scope

- [x] Correct isolated branch and exact start.
- [x] GSD fallback and required skills recorded.
- [ ] Fixture/replay/local-only certify checks.
- [ ] No real credential value or credentialed connector check.
- [ ] No live write/sweep, services, dependencies, defs, generic tools, destructive/admin/production action.
- [ ] Commit/push checkpoints complete; no PR/review.
