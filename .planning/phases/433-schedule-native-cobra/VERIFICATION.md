# Phase 433 Verification

Invocation: `issue-433-pi-sol-high-20260719T044819Z`; profile `Sol`; thinking `high`; exact start `ab1c79eede67fa87e1c6b808d6ddba0b27fcf00d`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [ ] Focused RED captured before production edits.
- [ ] Native schedule create/list/install/remove/help tree; legacy wrapper removed.
- [ ] All current local flags typed with repeated/bare/assigned behavior preserved where applicable.
- [ ] Bare/text/JSON/long/short/positional help parity.
- [ ] Trailing help, literal `--`, malformed/legal unknowns, invalid actions, and no action-discovery bypass.
- [ ] `uninstall`, `run`, and `history` remain invalid and effect-free under the current public contract.
- [ ] Strict first-operand ownership and unchanged project-root payload semantics.
- [ ] Cron/name/not-found/backend validation, context propagation, exact exit taxonomy, deterministic output, and one-envelope stdout/stderr behavior.
- [ ] Install/remove runtime effects covered with injected fakes/temp crontab only.
- [ ] Only schedule parser/dispatcher removed; dynamic connector parser remains.

## Gates

- [ ] Focused native schedule/router tests.
- [ ] Focused tests repeated (`-count=5`).
- [ ] Focused schedule/router race tests.
- [ ] Existing schedule CLI and schedule cron/manifest/render/select/config tests.
- [ ] Router and golden transcript focus; fixture unchanged.
- [ ] Exact-start differential for parser/output edge cases where deterministic.
- [ ] Full `go test ./internal/cli/...`.
- [ ] Full `go test ./internal/schedule/...`.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test -timeout 20m ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify` without external scheduler effects.
- [ ] `git diff --check`; no dependency/connector-definition/unrelated delta.

## CLI help/manual/website parity

- [ ] `pm help schedule`.
- [ ] Bare `pm schedule` exits 0 with contextual manual.
- [ ] `pm schedule --help`, `-h`, positional `help`, and JSON manual routes.
- [ ] Invalid action remains usage error, including trailing-help/literal cases.
- [ ] `docs/cli/schedule.md` generated parity checked; update or N/A with reason.
- [ ] Website CLI-reference and architecture schedule documentation checked/generated; update or N/A with reason.
- [ ] Generated/golden help artifacts checked and unchanged or explicitly updated.
- [ ] Completion discovery seam present; Phase 15 values remain deferred.
- [ ] Phase 11 schedule wizard and Phase 19 focused help/man churn remain deferred.

## Safety/scope/delivery

- [x] Exact branch/start and parent draft PR confirmed.
- [x] GSD doctor/list passed; unavailable programming-loop/manual fallback recorded.
- [x] Required CLI/testing/error/security/safety/context/docs/Cobra skills loaded.
- [ ] Temp roots/temp crontab/fake backends only; no external scheduler or live credentialed check.
- [ ] No optional services, dependencies, unrelated namespaces, or broad generated churn.
- [ ] Planning, RED, GREEN/refactor, and final checkpoints committed/pushed.
- [x] No PR/review will be created per user instruction.

Result: DRAFT; `verificationPassed=false` until the full declared `make verify` exits 0.
