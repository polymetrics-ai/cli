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

## Reusable foundation already present

The parent branch contains an earlier read-only Pi `AgentSession` control plane with strict parsing,
state/evidence validation, bounded concurrency, cancellation, lease ownership, offline command
registration, and a historical read-only #438 canary. Its prior green gates are useful regression
evidence, not overall completion proof.

Additional uncommitted adversarial remediation strengthens target evidence, state roots/leases,
SDK cleanup, lifecycle ownership, and extension coordination. It belongs to #473 and must be
finished test-first before integration.

## Remaining critical path

1. Finish #473 blockers: first-wins cancellation, authoritative lease acquisition across epoch
   cleanup races, root device/inode pinning, unsuppressed root failures, coherent state invariants,
   bounded epochs, and honest macOS threat documentation.
2. Open/review/integrate the #473 child PR into #472.
3. Dispatch #474-#477 concurrently in isolated worktrees.
4. Complete #478 parent/GitHub orchestration, #479 controller/UX integration, and #480 recovery/
   audit/reversible-cutover preparation.
5. Run #481 against #397/#438; only after it passes, activate legacy-shell deprecation, then run
   full local/CI gates, exact-head independent review, and the durable parent merge decision.

Overall verification remains false until those steps complete. #472 stays draft and cannot merge
without a fresh allowlisted human `approve-merge` decision for its exact verified head.
