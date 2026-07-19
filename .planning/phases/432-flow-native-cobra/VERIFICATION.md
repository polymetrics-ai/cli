# Phase 432 Verification

Invocation: `issue-432-pi-openai-codex-gpt-5.6-sol-high-20260719T034344Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `ec12c1729e0aaf233a853eff5c6291885f910b15`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [x] Focused RED captured before production edits (`undefined: newFlowCobraCommand`; flow cancellation contract independently passed).
- [x] Native flow plan/preview/run/list/status/help tree; legacy wrapper removed.
- [x] All current local flags typed with repeated/bare/assigned behavior preserved where applicable.
- [ ] Bare/text/JSON/long/short/positional help parity.
- [ ] Trailing help, literal `--`, malformed/legal unknowns, invalid actions, and no action-discovery bypass.
- [ ] Named-run/status first-operand ownership and unchanged directory semantics.
- [ ] Exact exit taxonomy, deterministic output, terminal sanitation, and one-envelope stdout/stderr behavior.
- [x] Cancellation, events, telemetry, checkpoint, and ledger contracts preserved in focused tests.
- [x] Only flow parser/dispatcher removed; dynamic connector parser remains.

## Gates

- [x] Focused native flow/router tests (`5.002s`; all flow CLI `5.742s`).
- [x] Focused tests repeated (`-count=5`: `27.066s`).
- [x] Focused flow/router race tests (CLI `60.885s`; flow `1.348s`).
- [x] Existing flow CLI and flow engine/action/checkpoint/event/telemetry tests (`0.301s`).
- [x] Router and golden transcript focus (`13.293s`); fixture unchanged.
- [x] Exact-start differential for parser/output edge cases (initial 180/200; focused 8-case RED; corrected 200/200 exact).
- [ ] Full `go test ./internal/cli/...`.
- [ ] Full `go test ./internal/flow/...` and relevant event/telemetry packages.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test -timeout 20m ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify`.
- [ ] `git diff --check`; no dependency/connector-definition/unrelated delta.

## CLI help/manual/website parity

- [ ] `pm help flow`.
- [ ] Bare `pm flow` exits 0 with contextual manual.
- [ ] `pm flow --help`, `-h`, positional `help`, and JSON manual routes.
- [ ] Invalid action remains usage error, including trailing-help/literal cases.
- [ ] `docs/cli/flow.md` generated parity checked.
- [ ] Website CLI-reference and architecture flow documentation checked/generated.
- [ ] Generated/golden help artifacts checked.
- [ ] Completion discovery seam present; Phase 15 values remain deferred.
- [ ] Phase 10 dashboards, Phase 11 create wizard, and Phase 19 focused help/man churn remain deferred.

## Safety/scope/delivery

- [x] Exact branch/start and parent draft PR confirmed.
- [x] GSD doctor/list passed; unavailable programming-loop/manual fallback recorded.
- [x] Required CLI/testing/error/security/safety/context/concurrency/docs/Cobra skills loaded.
- [ ] Temp manifests, temp roots, and in-memory fakes only; no action/external write or credentialed check.
- [ ] No optional services, dependencies, unrelated namespaces, or broad generated churn.
- [ ] Planning, RED, GREEN, and final checkpoints committed/pushed.
- [ ] No PR/review created per user instruction.

Result: pending.
