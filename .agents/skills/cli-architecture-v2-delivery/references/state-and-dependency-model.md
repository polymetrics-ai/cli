# State and dependency model

Use this algorithm for every CLI Architecture v2 planning, scheduling, audit, review, and integration
turn. Mutable issue numbers, edges, owners, heads, and queue entries belong in current GitHub and
parent orchestration artifacts—not this reference.

## Pin remote truth

1. Fetch the default branch, parent branch, and relevant worker branch without rewriting them.
2. Record the full object IDs for default, parent, worker, and PR head plus each relevant merge-base.
3. Query the parent issue, child issue, native dependencies, PR base/head, checks, and review state.
4. Inspect parent-branch code, tests, GSD phase artifacts, and integration commits.
5. Re-fetch before review, promotion, or handoff. Any changed head invalidates head-bound evidence.

Do not copy a PR-body head into current evidence without comparing it to the remote ref. PR prose,
issue checkboxes, local refs, and orchestration traces may be stale.

## Separate four graphs

- **Hierarchy** says an issue is a parent or child. It does not itself block execution.
- **GitHub-native dependencies** are authoritative scheduling edges when present.
- **Planning-only dependencies** are repository decisions that must be cited and checked for drift.
- **Write-scope collisions** serialize otherwise independent work touching central command, help,
  golden, docs, website, generated, or parent-state surfaces.

Compute readiness from all four. Never infer a dependency only from numbering or phase order.

## State definitions

| State | Evidence |
|---|---|
| `parent_branch_satisfied_at` | The named parent commit contains the required code, tests, artifacts, and integration evidence. Default-branch delivery may remain open. |
| `active_ready` | Dependencies and design gates are satisfied, write scope and owner are assigned, and no human gate blocks execution. |
| `dependency_blocked` | At least one verified dependency or design gate is unsatisfied. |
| `human_decision_blocked` | A product, dependency, security, or other explicit human decision is required. |
| `integrated_review_debt` | Implementation is already on the parent, but required exact-range review, disposition, or verification evidence is missing. |
| `deferred_by_human` | A human explicitly chose not to deliver the optional slice in this program. This is not implementation completion. |
| `default_branch_complete` | The required change is present on the default branch and issue delivery criteria are satisfied. |

Record the parent commit in the evidence for `parent_branch_satisfied_at`; do not put it in evergreen
skill prose.

## Legal transitions

- Evidence may move a slice from blocked to `active_ready`.
- Reviewed promotion may move a worker head to `parent_branch_satisfied_at`.
- Missing review evidence moves an integrated slice to `integrated_review_debt`, not back to
  implementation-ready.
- Only an explicit human decision may set or clear `deferred_by_human`.
- Only default-branch delivery may set `default_branch_complete`.
- Head drift returns head-bound review and verification to pending.

## Ready-queue recomputation

For each candidate, verify dependency gates, parent evidence, human gates, owner availability, and
write-scope overlap. Select only disjoint slices. Parent orchestration owns the resulting queue and
all promotions. A worker may report evidence but may not edit shared queue state unless explicitly
delegated.
