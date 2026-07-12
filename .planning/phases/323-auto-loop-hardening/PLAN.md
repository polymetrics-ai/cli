# Plan: Autonomous Delivery Control-Plane Hardening

Parent issue: #323
Incident provenance: #277
Parent branch: `fix/323-auto-loop-hardening`
Parent PR: #324 against `feat/pi-shepherd-loop` (draft)

## Outcome

Harden the existing `scripts/pi-shepherd-loop.sh` delivery path in place so it becomes fail-closed,
transactional, restart-safe, independently validated, and cost-bounded. The existing launcher,
prompts, Pi agents, and trace workflow remain the compatibility surface; helper code may enforce
their contracts but must not replace them with a parallel delivery system. The final parent PR
remains draft and human-gated until every child issue is integrated, replay/fault-injection
verification is green, and automated review coverage is complete.

## Baseline and isolation

- The hardening branch originally forked from `origin/main` at `cab8f3df`. PR #324 was then based on
  `feat/pi-shepherd-loop`, and commit `2b40ba52` was merged normally into this branch on 2026-07-12.
  This preserves the implementation used by the Twenty run and the Phase 0 replay/safety history
  without rebasing or force-pushing either published line.
- `scripts/pi-shepherd-loop.sh` is the canonical supervised implementation. Hardening work updates
  that entrypoint and its existing contracts; `scripts/claude-auto-loop.sh` and
  `scripts/pi-auto-loop.sh` remain compatibility entrypoints behind the same migration fuse.
- The Twenty connector branch, PR #285, connector bundles, and dirty run checkout are out of scope.
- Every mutating child worker receives one issue, one branch from this parent branch, one isolated
  worktree, one bounded write scope, and one handoff.

## GSD execution record

- `PATH="$HOME/.nvm/versions/node/v24.13.1/bin:$PATH" scripts/gsd doctor` passes.
- The documented `scripts/gsd prompt programming-loop ...` route is absent from the current
  69-command adapter registry and fails as an unknown command.
- Fallback preflight used the installed `gsd-programming-loop` helper in dry-run agent mode. The
  live lifecycle remains the repository's `gsd-universal-runtime-loop.md`, with strict red/green
  evidence on every behavior-changing child issue.
- A later helper run misclassified this shell/control-plane phase as UI work. Its generated output
  was preserved in a named stash and excluded from implementation; no generated file was deleted.
- Required skills loaded for parent planning: `github-issue-first-delivery`,
  `gsd-programming-loop`, `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`,
  `golang-concurrency`, `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-naming`.

## Delivery slices

| Slice | Issue / branch | Child outcome | Depends on | Primary write scope |
|---|---|---|---|---|
| 0 | #325 / `fix/325-agentloop-characterization` | Sanitized replay oracle and default-deny migration fuse | parent PR #324 | fixture/replay/safety core, `cmd/loopctl`, driver guards, tests |
| 1A | #326 / `fix/326-controller-fencing` | Controller fence, durable lease, heartbeat, turn/session limits | #325 | minimal shared model, controller store/manager/heartbeat/turn, read-only status |
| 1B | #339 / `fix/339-controller-takeover` | Authenticated process guardian, exact sessions, fenced takeover and cleanup | #326 | controller process/takeover/adapter core, CLI and driver adapters |
| 2A | #327 / `fix/327-transactional-state` | Atomic CAS journal, events, create-only checkpoints, recovery | #339 | state/model/checkpoint core and schemas |
| 2B | #335 / `fix/335-authority-terminal` | Contract manifest, durable authority/HALT/RETRY, human decisions, terminals | #327 | contract/authority/terminal core and schemas |
| 3 | #328 / `fix/328-gate-tickets` | Gate-closure compiler and immutable one-stage tickets | #335 | `internal/agentloop/gate/**`, validation contracts |
| 4A | #329 / `fix/329-worker-leases` | Worker/worktree leases, quiescence, complete scope audit | #328 | worker/scope core and schemas |
| 4B | #336 / `fix/336-capability-enforcement` | Production mutation guard and controller-only publication | #329 | capability/publish core and worker adapters |
| 5 | #330 / `fix/330-atomic-validation` | Atomic read-only Shepherd transaction and moved-evidence rejection | #336 | validator core, prompt and adapter |
| 6A | #331 / `fix/331-github-outbox` | Transition-bound idempotent GitHub outbox | #330 | outbox/closed GitHub adapter and schemas |
| 6B | #337 / `fix/337-merge-attestation` | Exact-head one-shot attestation and ratification boundary | #331 | attestation/merge guard and required check |
| 7A | #332 / `fix/332-typed-trace` | Bounded redacted event-derived trace projection | #337 | event/trace/redaction core and wrapper |
| 7B | #338 / `fix/338-dispatch-session-contracts` | Typed dispatch/HANDOFF, session compaction and usage de-duplication | #332 | dispatch/handoff/usage core and Pi adapter |
| 8 | #333 / `fix/333-deterministic-watchers` | No-model watchers, deterministic review packs, risk budgets | #338 | watcher/review/budget core and policy |
| 9 | #334 / `fix/334-fault-canary` | Historical replay, fault injection, shadow execution, merge-disabled canary | all prior | replay/fault suites and final evidence |

All fifteen issues are native sub-issues of #323. Phases 1, 2, 4, 6, and 7 were split into A/B
slices after read-only architecture review found their original scopes were not honestly PR-sized.
Dependency order is intentionally conservative because phases 1-8 share control-plane contracts;
parallelism is reserved for read-only design/review or demonstrably disjoint test work.

## Existing implementation alignment checkpoint

- The Phase 0 replay oracle and migration fuse now recognize all three existing launchers,
  including `scripts/pi-shepherd-loop.sh`, and deny run/resume before state or subprocess effects.
- Only the Shepherd validator default moves to `openai-codex/gpt-5.6-sol` with `high` reasoning.
  The orchestrator and every worker remain on their existing models.
- The default validator path performs an exact local model-catalog preflight before state creation
  or the first orchestrator turn. Pi 0.80.6 is the verified minimum with the Sol model entry;
  unsupported runtimes fail closed instead of silently downgrading.

## TDD and checkpoint policy

For every behavior-changing child issue:

1. Create/update its `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `PROMPTS.md`, and run state.
2. Capture an expected failing characterization or regression test before production edits.
3. Commit the plan/red checkpoint when useful; implement the smallest green slice.
4. Run targeted tests and race tests, then the accumulated historical replay suite.
5. Push only coherent green checkpoints to the child branch.
6. Open a stacked sub-PR against this parent branch with `Refs` for both issues.
7. Integrate only after local/remote gates and automated review coverage satisfy the stacked PR
   workflow. Parent-PR fallback coverage is recorded when the sub-PR reviewer skips a non-default
   base.

## Parent verification

The parent cannot become human-ready until all applicable commands pass:

```bash
PATH="$HOME/.nvm/versions/node/v24.13.1/bin:$PATH" scripts/gsd doctor
go test ./internal/agentloop/...
go test -race ./internal/agentloop/...
go test ./cmd/loopctl/...
scripts/tests/auto-loop-control.sh
make verify
```

Phase 9 additionally proves controller takeover, child revocation, stale validator rejection,
head/lease movement rejection, single-winner CAS, duplicate outbox suppression, trace identity
uniqueness, bounded child-session references, and risk-budget escalation.

## Human gates

- Merging the parent PR to `main`.
- New dependencies or quality-gate reductions.
- Authentication-scope changes or auth refresh.
- Repository ruleset or merge-principal changes.
- Secret access, credentialed provider checks, destructive external actions, or production deploys.
- Any generic shell, generic HTTP write, generic SQL write, or unrestricted raw API capability.
