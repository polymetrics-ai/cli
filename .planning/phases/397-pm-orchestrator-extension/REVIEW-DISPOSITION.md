# PM Orchestrator Extension Local Codex Review Disposition

Review range: `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f...3c88fc78062ba0a3437f79bc88c395286c228c65`
Reviewer: fresh-context project `pm-reviewer`, local Codex Sol/xhigh, read-only
Status: correction in progress; exact-head re-review required

| ID | Severity | Disposition | Reason / action |
|---|---|---|---|
| F1 | high | accepted_with_modification | PR #495 will remove the unconditional unavailable Pi command example, extend focused validation to the adapter, and make PR #493's subsequent routing reconciliation a hard pre-worker gate. `AGENTS.md` and required-skills routing remain untouched because captain explicitly assigned those paths to PR #493; PR #495 must not duplicate or absorb them. |
| F2 | medium | deferred_out_of_scope | The three Gong definitions and validator/runtime behavior are byte-identical to current `origin/main` commit `873cd7b251f70c4a35a607a0d4e86051ea0fbd15` (`git diff --quiet origin/main..3fd63fbe --` on all cited files). This is a pre-existing Gong product defect, not introduced or worsened by Wave 1 or the docs/workflow extension. Original scope forbids product behavior beyond conflict preservation. Focused follow-up: https://github.com/polymetrics-ai/cli/issues/497. |
| F3 | medium | accepted | Make the current review schema versioned and conditional: canonical pending/blocked Shepherd records need no invented verdict; completed statuses require one. Preserve legacy bot and local-Codex shapes read-only. Add pending/clean/blocked/historical fixtures to focused validation. |
| F4 | medium | accepted | Add a required default correction budget (4), per-range counters, and human-block transition to the parent contract, `/pm-orchestrate`, local review workflow, state schema/spec, and focused validation. |
| F5 | medium | accepted | Update PR #495 title/body after correction commits to distinguish historical `3fd63fbe...` evidence from the new exact head and mark/reconcile current review, Shepherd, and CI state truthfully. Body-only updates do not alter Git head. |

## Residual human/program gates

- PR #493 must reconcile its owned `AGENTS.md`, required-skills routing, task-skill matrix, skill, and Makefile changes to the canonical PM route after Wave 1 lands and before another CLI Architecture v2 worker starts.
- The Gong operation direct-read defect is tracked at https://github.com/polymetrics-ai/cli/issues/497; no credentialed or live call is needed to reproduce its local preflight failure.
- Parent PR #438 and PR #495 remain draft/human-only and unmerged.
