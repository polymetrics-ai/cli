# Parent Issue Orchestrator Contract

Use this contract when a parent issue owns multiple sub-issues and stacked PRs. The orchestrator is
the single owner for shared parent artifacts, merge arbitration, automated review coverage routing,
and final human approval readiness.

## Required Input

The parent issue must provide:

- objective
- background
- acceptance criteria
- parent branch name
- parent PR URL or explicit blocker
- sub-issue roster with dependencies and branch names
- verification commands
- human gates
- source links

The orchestrator must also read:

- `AGENTS.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`

## Responsibilities

The orchestrator owns:

- creating or confirming the parent branch from `main`
- creating the parent PR to `main` before sub-issue execution
- creating a deliberate parent seed commit when GitHub needs a diff to open the parent PR
- maintaining the parent issue status and orchestration state ledger
- selecting sub-issues that can run in parallel without write-scope collisions
- spawning or assigning worker agents with bounded prompts that name the `/gsd ...` or `scripts/gsd prompt ...` command path and required Go/design skills from `required-skills-routing.md`
- receiving worker handoffs
- deciding whether a sub-PR can merge into the parent branch
- requesting or observing parent PR CodeRabbit coverage after integrated batches
- routing Copilot backup review when CodeRabbit is blocked by rate limits or unavailability
- launching or assigning automated review disposition work
- declaring final parent PR readiness for human approval

Worker agents own implementation for one assigned sub-issue. Workers do not own shared parent issue
comments, parent PR bodies, parent branch pushes, or default sub-PR merge decisions unless the
orchestrator explicitly delegates those actions.

## State Machine

Use these states in the parent issue, parent PR, or durable state ledger:

- `planned`: parent issue exists, but parent branch or parent PR is not ready.
- `parent_pr_open`: parent branch and draft parent PR exist.
- `worker_ready`: a sub-issue has complete inputs and no blockers.
- `worker_in_progress`: a worker is implementing the sub-issue.
- `sub_pr_open`: worker opened a sub-PR against the parent branch.
- `sub_pr_green`: sub-PR local and remote checks are green.
- `sub_pr_reviewed`: automated review coverage exists on the sub-PR, or an explicit fallback is
  planned.
- `provisionally_integrated`: sub-PR merged into parent branch, but parent PR review coverage is
  still pending.
- `parent_review_pending`: automatic parent PR review is running, coverage is waiting for the
  parent PR to leave draft, or an allowed fallback review route is recorded for an integrated batch.
- `parent_review_clean`: integrated batch has no unresolved actionable automated review findings.
- `final_verification`: all sub-issues are integrated and full parent verification is running.
- `ready_for_human`: parent PR is ready, but merge to `main` still needs human approval.
- `blocked`: a human gate, failed verification, review blocker, or dependency blocks progress.
- `complete`: parent PR merged to `main` and closing issue references have landed.

## Parallelism Rules

Sub-issues may run in parallel only when all of these are true:

- dependency order permits it
- write scopes are disjoint
- each worker has one primary issue
- each worker has a bounded branch and PR base
- shared parent artifacts are orchestrator-owned
- human gates are not crossed

When file ownership is unclear, run the sub-issues sequentially or split them further.

## Merge Policy

A sub-PR may merge into the parent branch only when:

- it targets the parent branch
- it uses `Refs #<sub-issue>` and `Refs #<parent-issue>`, not closing keywords
- targeted and issue-level verification pass
- CI checks pass or an infrastructure blocker is recorded
- automated review findings on the sub-PR are resolved, or the parent PR fallback path is recorded
- the diff is within the sub-issue scope
- no requested-changes review is open
- no human gate is triggered

If CodeRabbit skips a non-`main` sub-PR, merge into the parent branch is only provisional. The
orchestrator must observe automatic parent PR review, or record an allowed fallback route such as
Copilot backup or human review, for the integrated commit range before marking that sub-issue
review-complete.

The parent PR into `main` always requires human approval.

## Automated Review Coverage Record

For every sub-issue, record:

- sub-issue number
- sub-PR URL
- parent PR URL
- base branch
- head branch
- head SHA
- reviewed commit or commit range
- primary route: `coderabbit_auto`, `coderabbit_auto_incremental`,
  `coderabbit_manual_fallback`, `copilot_backup`, `human`, or `blocked`
- coverage route: `sub_pr`, `parent_pr_fallback`, `copilot_backup`, or `blocked`
- fallback route: `copilot_backup`, `human`, or `none`
- review status: `pending`, `clean`, `comments_addressed`, `skipped`, or `blocked`
- disposition summary URL or comment

## Output Requirements

The orchestrator must leave a reviewer able to answer:

- which sub-issues were launched
- which worker owned each sub-issue
- which branches and PRs were used
- which checks ran
- which automated review route covered each sub-issue
- why any sub-issue was deferred or blocked
- whether the parent PR is ready for human approval
