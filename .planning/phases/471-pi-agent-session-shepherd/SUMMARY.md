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

## Aggregate implementation now present

The #479 production-matrix branch contains the capabilities originally decomposed across #473-#478:
the durable control plane, policy scheduler, scoped in-process `AgentSession` runtime, isolated
workspace/Git lifecycle, GitHub decision broker, parent orchestration, and production composition.
All 17 production rows functionally pass at code head `91692415`. This branch-local aggregation
does not claim that every independent child issue or stacked PR is closed; reconcile those records
as part of non-default parent integration.

## Remaining critical path

1. Run the complete Shepherd inventory in fresh CI outside the managed sandbox.
2. Integrate the verified #479 aggregate into the non-default parent branch and reconcile the
   #473-#478 issue/stacked-PR records without replaying already aggregated code.
3. Complete #480 recovery/audit/reversible-cutover preparation.
4. Run #481 against #397/#438; only after it passes, activate legacy-shell deprecation, then run
   full local/CI gates, exact-head independent review, and the durable parent merge decision.

Overall verification remains false until those steps complete. #472 stays draft and cannot merge
without a fresh allowlisted human `approve-merge` decision for its exact verified head.
