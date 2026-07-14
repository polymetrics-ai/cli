# Issue 389 Shepherd Hardening Plan

## Objective

Make `shepherd supervise --config <path> --issue <N> --context <path>` the single issue-scoped
entry point from validated intake to the final human gate. The supervisor must fail closed when its
GSD runtime contract, issue identity, retry authority, child lifecycle, or completion proof is
ambiguous.

## Required workflow and skills

- GSD command: `scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`
- Skills: `gsd-programming-loop`, `github-issue-first-delivery`, `golang-how-to`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, `golang-error-handling`, `golang-safety`, `golang-testing`,
  `golang-observability`, and repo-local `.pi/skills/go-implementation`.
- Parent delivery: issue #389 is stacked under parent #372 and parent PR #390. The final merge to
  `main` remains human-gated.

## Slice 1 — Runtime prompt/tool contract admission

- RED: a `plan-milestone` prompt advertises `gsd_resume` while its registry omits that tool.
- GREEN: Shepherd discovers the active GSD resource root, derives advertised and allowed tools,
  and rejects any mismatch before Pi starts with `runtime_contract_mismatch`.
- Refactor: keep compatibility checks version-qualified and side-effect free; retain the existing
  exact 1.11.0 headless patch.

## Slice 2 — Canonical issue-scoped GSD identity and bootstrap

- RED: two issue identities attempt to share one project root/controller state; an existing issue
  is restarted with a changed branch, base, root, or initial head.
- GREEN: persist immutable issue, parent issue, branch, base, GSD project root, initial head, GSD
  version, and controller generation. Materialize bounded issue-local GSD planning artifacts
  transactionally and adopt an exact existing binding idempotently.
- Refactor: preserve GSD's project-root database as workflow truth and Shepherd SQLite as controller
  truth; never merge databases between issues.

## Slice 3 — Durable retry and signal/subagent lifecycle

- RED: repeated process restarts exceed the configured unit budget; a signal leaves a persisted
  subagent `running` without a live process or resumable session; nested work produces no parent
  progress.
- GREEN: count attempts durably by issue/unit/generation/head, reconcile orphaned subagent records
  to `interrupted`, and project nested activity into the existing heartbeat at no more than 15-second
  gaps.
- Refactor: cancellation remains context-owned, bounded, idempotent, and free of chain-of-thought.

## Slice 4 — Exact completion proof

- RED: a zero-exit unit lacks its expected artifact, leaves an active child, fails to advance GSD
  SQLite/query state, or runs against a stale Git head.
- GREEN: success requires canonical state advancement, exact-head continuity, expected artifacts,
  no live/unreconciled children, governed model evidence, a clean scoped checkpoint, and a current
  lease.
- Refactor: classify failures with typed evidence instead of treating every failure as a generic
  runtime error.

## Slice 5 — One-command issue supervision

- RED: a deterministic fake GSD workflow requires operator-selected commands or dispatches the same
  unit twice after restart.
- GREEN: `supervise` bootstraps/adopts one issue, selects only the canonical next unit, routes
  planning/validation to GPT-5.6 Sol/high and execution to GPT-5.5/high, applies bounded recovery,
  emits live progress, and stops at blocked/human/final merge gates.
- Refactor: keep policy deterministic and keep external effects behind existing narrow ports.

## Verification and checkpoints

1. Focused test after each RED/GREEN slice.
2. `go test ./...`
3. `go test -race ./...`
4. `go vet ./...`
5. `go build ./cmd/shepherd`
6. `make verify` in the nested module.
7. Root `go list ./...` proves the Shepherd module remains excluded from `pm`.
8. Root `go test ./...`, `go build ./cmd/pm`, and `make verify` when shared files change.

Commit and push each coherent green checkpoint to `feat/389-autonomous-shepherd`; never push or
merge to `main`.
