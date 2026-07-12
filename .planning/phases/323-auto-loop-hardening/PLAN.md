# Plan: Autonomous Delivery Control-Plane Hardening

Parent issue: #323
Incident provenance: #277
Parent branch: `fix/323-auto-loop-hardening`
Parent PR: pending creation against `main`

## Outcome

Replace the prompt-enforced connector delivery loop with a fail-closed, transactional,
restart-safe, independently validated, and cost-bounded control plane. The final parent PR remains
draft and human-gated until every child issue is integrated, replay/fault-injection verification is
green, and automated review coverage is complete.

## Baseline and isolation

- The branch starts at `origin/main` commit `cab8f3df`, which contains the original Claude driver
  and Shepherd validator from PR #276.
- The post-merge loop commits used by the Twenty run (`0087216c` through `2b40ba52` on local branch
  `feat/pi-shepherd-loop`) are evidence/baseline inputs. They are not assumed to be on `main` and
  must be adopted deliberately by the characterization child issue if still required.
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
- Required skills loaded for parent planning: `github-issue-first-delivery`,
  `gsd-programming-loop`, `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`,
  `golang-concurrency`, `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-naming`.

## Delivery slices

| Phase | Child outcome | Depends on | Primary write scope |
|---|---|---|---|
| 0 | Sanitized characterization/replay harness; irreversible-action kill switch defaults off | parent PR | `scripts/tests/**`, test fixtures, baseline loop adapters, phase artifacts |
| 1 | Singleton controller generations, leases, process-group ownership, heartbeat, takeover, cleanup, exact session binding, per-turn cap | 0 | `internal/agentloop/controller/**`, `cmd/loopctl/**`, driver adapter |
| 2 | Versioned CAS state/events/checkpoints, durable RETRY/HALT, signed human decisions, terminal model | 1 | `internal/agentloop/state/**`, state schemas, checkpoint adapter |
| 3 | Gate-closure compiler and immutable one-stage tickets | 2 | `internal/agentloop/gate/**`, validation contracts |
| 4 | Worker/worktree leases, production capabilities, quiescence, scope audit | 3 | `internal/agentloop/worker/**`, worker adapter/contracts |
| 5 | Atomic read-only Shepherd transaction and moved-evidence rejection | 4 | `internal/agentloop/validator/**`, validator prompt/driver |
| 6 | Transition-bound idempotent GitHub outbox and exact-head merge attestation | 5 | `internal/agentloop/outbox/**`, guarded GitHub adapter/workflow |
| 7 | Event-derived trace/UI, typed dispatch/HANDOFF, redaction, canonical usage/session IDs | 6 | `internal/agentloop/trace/**`, trace CLI/scripts, schemas/prompts |
| 8 | No-model CI/worker/human watchers, deterministic review packs, risk budgets | 7 | `internal/agentloop/watcher/**`, review routing, driver policy |
| 9 | Historical replay, fault injection, shadow execution, merge-disabled canary | 0-8 | replay/fault suites, final phase artifacts only |

Child issue numbers, branch names, worker directories, and PR URLs will be filled after native
sub-issues are created. Dependency order is intentionally conservative because phases 1-8 share
control-plane contracts; parallelism is reserved for read-only design/review or demonstrably
disjoint test work.

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
