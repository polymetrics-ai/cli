# Summary

Phase: `471-pi-agent-session-shepherd`
Status: `worker_ready` (#480); #481 dependency-blocked

## Completed in the replacement pivot

- #471 is rewritten as the authoritative autonomous in-process Shepherd parent.
- Draft parent PR #472 targets `main` from `feat/471-pi-agent-session-shepherd`.
- Child issues #473-#481 define six dependency waves with isolated write scopes.
- Abandoned Go/tmux issues #372/#389/#470 are closed `not_planned`; draft PRs #390/#456 are closed
  unmerged. Historical branches/worktrees are preserved.
- Role routing is `openai-codex/gpt-5.6-sol`: implementation/correction `high`, all other roles
  `xhigh`.
- The parent workflow, phase context/plan, prompt contract, docs, and human-decision protocol now
  describe a complete autonomous replacement rather than a read-only companion.

## Aggregate implementation available on the child branch

The #479 production-matrix branch contains the capabilities originally decomposed across #473-#478:
the durable control plane, policy scheduler, scoped in-process `AgentSession` runtime, isolated
workspace/Git lifecycle, GitHub decision broker, parent orchestration, and production composition.
The original 17-row matrix is `91692415`, current production code is `78708cbe`, the deterministic
Pi-family CI correction is `a594be98`, and child evidence is `d895dc38`. This branch-local
aggregation does not claim current review/CI readiness or that every independent child issue or
stacked PR is closed.

## PR #489 CI repair

The reconciled parent was published at `45c27b9d`, and child PR #489 now targets that exact
non-default branch. Its first ordinary-host run passed nine checks and exposed two bounded release
failures: the Shepherd workflow lacked an exact Go setup for its real-Go fixture, and the stack had
not yet inherited `golang.org/x/text v0.39.0` from current `main`. The current main baseline
`873cd7b2` is merged into the parent at `383fcf93`; the child CI pins Go `1.25.12` at `52866972`.
The focused fixture, exact Pi family, strict TypeScript, YAML, and module-integrity checks pass
locally. The complete inventory and `govulncheck` still require their ordinary-host CI rerun because
the managed sandbox blocks `/bin/ps` and external advisory-database DNS respectively.

## Reconciled state after #491

- Clean local, remote, and GitHub parent head: `c3f4f683e60ac52bcedae04b2e9448e4523b5234`.
- #473-#479 and #490: integrated into the non-default parent branch. Their issues remain open until
  the default-branch parent lands; previously closed #473/#474/#476/#477 were reopened.
- #480: `worker_ready`; its dependency #479 is integrated and all retained child branches/worktrees
  are historical rather than active authority.
- #481: `dependency_blocked` on #480. PR #438 is open/draft at `21d195af`, but currently conflicts
  with `main`; the #481 canary must reconcile that exact state read-only and must not merge #438.
- `/pm-shepherd status --issue 471` reported no persisted run. One validated ignored schema-2 plan
  will start #480 and then #481 in persistent child worktrees; after state exists, only `resume` is
  allowed.
- The user-selected review route is one independent Codex 5.6-sol xhigh parent round, not
  Claude/Copilot and not human approval.

## Remaining critical path

1. Commit/push this reconciliation plan and start the single durable Shepherd run for #480/#481.
2. Integrate each exact-head child only after focused/full Shepherd gates, CI, and one bounded
   review/correction round.
3. Freeze the final parent SHA, review the unreviewed range and cross-child seams in four parallel
   domains, and apply at most one concrete blocker correction pass.
4. Run the strict Pi-family/typecheck/provenance/RPC gates and the complete sequential Shepherd suite
   exactly once on the final head, then require exact-head GitHub CI green.
5. Update #471/#472, mark #472 ready, and request only the human merge of the exact SHA.

Overall verification remains false until those steps complete. #472 cannot merge without the final
human gate and the orchestrator never merges it to `main`.
