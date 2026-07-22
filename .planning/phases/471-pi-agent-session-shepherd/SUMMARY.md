# Summary

Phase: `471-pi-agent-session-shepherd`
Status: `parent_pr_open`

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

## Remaining critical path

1. Publish/fetch the reconciled parent branch, then push/open the #479 child PR against that exact
   non-default base.
2. Run the complete Shepherd inventory in fresh CI, complete exact-head internal Codex review, and
   obtain repository-policy review coverage or an allowed recorded fallback.
3. Integrate #479 into the non-default parent and reconcile #473-#478 lifecycle records.
4. Complete #480 recovery/audit/reversible-cutover preparation.
5. Run #481 against #397/#438; only after it passes, activate legacy-shell deprecation, then run
   full local/CI gates, exact-head independent review, and the durable parent merge decision.

Overall verification remains false until those steps complete. #472 stays draft and cannot merge
without a fresh allowlisted human `approve-merge` decision for its exact verified head.
