# Phase 479: production Shepherd MVP

## Objective

Deliver one working autonomous Shepherd path that owns intake, dependency-aware parallel child
implementation, isolated worktrees, shell verification, commit/push/stacked PR publication,
independent exact-head review and correction, child integration, crash-safe resume, and an exact-head
human parent-merge wait. Shepherd never merges the parent PR into the default branch.

This phase expands the bounded #479 controller MVP to the complete 17-row production contract from
the preflight matrix. None of the rows below are deferred as post-MVP hardening.

## Process contract

- Active issue: #479; parent issue: #471; parent branch: `feat/471-pi-agent-session-shepherd`.
- Implementation branch: `feat/479-shepherd-production-matrix`.
- Parent integration is local until GitHub authentication is healthy. Never push or merge `main`.
- Strict RED -> GREEN -> REFACTOR. One consolidated blocker-only review follows GREEN; one fix pass.
- The repository-local `scripts/gsd` adapter was checked. `doctor` and `list` pass, but
  `scripts/gsd prompt programming-loop ...` reports the command unavailable. This phase therefore
  records the permitted manual-GSD fallback and follows the same lifecycle explicitly.
- Required skills: `gsd-programming-loop`, `gsd-workstreams`,
  `github-issue-first-delivery`, `architecture-patterns`, `javascript-testing-patterns`.
- Implementation workers use `gpt-5.6-sol/high`; orchestration and final review use xhigh.

## Frozen architecture

```text
Pi /pm-shepherd
  -> AutonomousShepherdController (single coordinator generation + CAS state)
     -> RepositoryPlanIntake (deep canonical plan + digest)
     -> dependency scheduler (dependencies, canonical scope collisions, capacity, idle reason)
     -> ProductionChildLifecycle
        -> WorkspaceAdapter + GitAdapter (isolated worktree and mutation lease)
        -> embedded AgentSession implementation/correction roles
        -> BoundedVerificationRunner (fixed executable+argv, no shell)
        -> typed commit/push and authoritative reconciliation
        -> GitHubParentOrchestrator (issue, PR, evidence, review, integration)
     -> GitHubDecisionBroker (bound parent merge request/observation)
     -> AutonomousEffectJournal (prepared -> observed -> applied)
     -> RecoveryBarrier (reconcile every uncertain effect before scheduling)
```

The model never receives generic shell, Git, GitHub, or merge-main capabilities. Verification
commands are plan-owned structured executable/argv records. External effects are adapter-owned,
bounded, cancellable, idempotency-keyed, and reconciled against authoritative state.

## Acceptance matrix

- [ ] 1. Complete intake -> parallel implementation -> PR -> review -> integration -> human wait.
- [ ] 2. Dependency order, scope collisions, concurrency caps, and durable idle reasons.
- [ ] 3. Disjoint worktrees coexist and every path releases only its own mutation lease.
- [ ] 4. Parent movement fences stale siblings; refresh/reclaim/rebase forces reverify/rereview.
- [ ] 5. Durable retry/correction budgets stop deterministically when exhausted.
- [ ] 6. Every external-effect checkpoint survives crashes without duplicate effects.
- [ ] 7. Stale, unauthorized, bot, edited, duplicate, and ambiguous replies fail closed.
- [ ] 8. Stop at every stage aborts and joins accepted work before state/lease release.
- [ ] 9. Stop/shutdown races, stale-generation fencing, and sibling abort are deterministic.
- [ ] 10. Commit/push/PR timeout reconciliation prevents duplicate publication.
- [ ] 11. Resume retains immutable ownership; changed plan/base/scope fails before mutation.
- [ ] 12. Findings have dispositions; head movement requires a clean exact-head rereview.
- [ ] 13. Dirty/scope-escaped/wrong-head/draft/untrusted-CI/prose-only evidence fails closed.
- [ ] 14. Parent head movement invalidates approval; no parent-to-main merge effect exists.
- [ ] 15. Traversal, symlink, controls, hostile payloads, timeout, output, and cancel are bounded.
- [ ] 16. Help, bare, invalid, status, and initialization-stop behavior is deterministic.
- [ ] 17. Top-level read-only children are rejected; internal read-only roles cannot integrate.

## Collision-safe implementation workstreams

| Lane | Exclusive ownership | Contract consumed |
|---|---|---|
| A: durable autonomy | new production state/effect/recovery modules and their tests | frozen DTOs in `autonomous-production-contract.ts` |
| B: workspace/verification | new bounded verification and production workspace/Git lifecycle modules and tests | state/effect ports and existing Git/Workspace adapters |
| C: GitHub/human/review | new concrete GitHub transport/gate/review composition and tests | existing orchestrator, decision, evidence and review ports |
| D: integrator | controller, local intake upgrade, extension/index wiring, trajectory matrix, docs | exported A-C ports |

Workers must not edit another lane's files or redefine shared DTOs. Contract changes return to the
integrator, update the RED scaffold first, and then unblock dependent lanes.

## Checkpoints

1. Plan, frozen contracts, and one compiling 17-row RED matrix.
2. Durable state, effect journal, generation fencing, retry/correction and recovery barrier.
3. Isolated worktree lifecycle, bounded shell verification, commit/push reconciliation.
4. GitHub issue/PR/evidence/review/integration and exact-bound human gate.
5. Controller composition, stale-parent refresh, stop/join, help/status/initialization behavior.
6. Focused tests, full Shepherd suite, strict TypeScript, offline Pi RPC, diff/scope checks.
7. One consolidated blocker-only independent review and one correction pass.
8. Integrate the exact verified head into the local parent branch; leave parent/main human-gated.

## Verification commands

```bash
node --test .pi/extensions/shepherd/*.test.ts
npx --yes --package typescript tsc --noEmit --strict --target es2022 --module nodenext \
  --moduleResolution nodenext --skipLibCheck .pi/extensions/shepherd/*.ts
pi --list-extensions
git diff --check
```

GitHub-backed tests use deterministic transports unless a healthy authenticated `gh` session is
available. Tokens are never printed, stored, or passed in prompts.
