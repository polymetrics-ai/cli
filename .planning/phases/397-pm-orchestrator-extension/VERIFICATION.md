# Issue #397 PM Orchestrator Extension Verification

Status: pending
`verificationPassed`: false

## Identity and scope

- [x] Isolated worktree verified.
- [x] Additive branch/PR #495 starts at reviewed head `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`.
- [x] Current main `873cd7b251f70c4a35a607a0d4e86051ea0fbd15` and parent `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f` pinned after fetch.
- [x] PR #493 head/file ownership inspected at `e21e56339390c5e1946eb4cfaf276eb80a889f29`.
- [ ] No PR #493-owned path changed.
- [ ] No #408/TUI/product/runtime/dependency change.

## Forward route

- [ ] `/pm-orchestrate` is the required owner for parent/stacked work when `programming-loop` is absent.
- [ ] PM route preserves PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE.
- [ ] State reconciliation, worker isolation, bounded correction, machine contracts, credential safety, and human merge authority remain explicit.
- [ ] Fresh-context local Codex review binds exact base/head and dispositions every finding.
- [ ] Any head change requires re-verification and fresh local Codex re-review.
- [ ] Shepherd is independent and required after review/before integration.
- [ ] Claude/Copilot are not required, requested, or fallback coverage in current/future PM route.
- [ ] Legacy Claude/Copilot docs/agent remain discoverable only with migration/deprecation pointers.
- [ ] Historical phase evidence remains unchanged.

## Focused validation

- [ ] RED captured before canonical guidance changes.
- [ ] `scripts/tests/pm-orchestrator-contract.sh`.
- [ ] `scripts/tests/pi-model-routing.sh`.
- [ ] YAML/JSON parse checks.
- [ ] PR #493 changed-path disjointness.
- [ ] no dependency delta.

## Full gates

- [ ] `gofmt -w cmd internal`
- [ ] `git diff --exit-code -- cmd internal`
- [ ] `git diff --check`
- [ ] `go vet ./...`
- [ ] `go test -timeout 20m ./...`
- [ ] `go build ./cmd/pm`
- [ ] `go mod verify`
- [ ] `go mod tidy -diff`
- [ ] `make verify`

## Review and delivery

- [ ] Fresh exact-head local Codex review clean.
- [ ] Every finding dispositioned; changed heads re-reviewed.
- [ ] Exact-head Shepherd/trajectory validation passes.
- [ ] Branch pushed normally without force.
- [ ] PR #495 title/body updated with extension scope and exact head.
- [ ] PR #495 remains draft; PR #438 remains draft/unchanged.
- [ ] No Claude/Copilot requested or counted.
- [ ] Required PR checks green at exact head.
