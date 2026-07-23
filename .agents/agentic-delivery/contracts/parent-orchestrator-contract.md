# Parent Issue Orchestrator Contract

Use this contract when a parent issue owns multiple sub-issues and stacked PRs. The orchestrator is
the single owner for shared parent artifacts, merge arbitration, exact-head verification, local
Codex review, independent Shepherd validation, and final human approval readiness.

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
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/local-codex-review-loop.md`
- `.agents/agentic-delivery/workflows/shepherd-validator.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`

## Activation

When triggered, the orchestrator is the active execution owner, not an advisory reviewer. It creates
or confirms parent branch/PR state, assigns workers, records state, arbitrates sub-PR merges, and
drives review coverage until the parent issue is human-ready or blocked. Runtime adapters must
invoke this contract; they must not fork or weaken its GSD, TDD, review, compact-mode, or human-gate
rules.

Triggers:

- user references a parent issue with sub-issues
- user references stacked PRs, parent branch, or parent PR
- parent PR is missing or lacks exact-head local review and Shepherd coverage
- sub-PR merge arbitration is needed
- local Codex review or Shepherd trajectory coverage blocks integration
- remaining sub-issues are ready or need dependency scheduling

## Responsibilities

The orchestrator owns:

- creating or confirming the parent branch from `main`
- creating the parent PR to `main` before sub-issue execution
- creating a deliberate parent seed commit when GitHub needs a diff to open the parent PR
- maintaining the parent issue status and orchestration state ledger
- selecting sub-issues that can run in parallel without write-scope collisions
- spawning or assigning worker agents with bounded prompts that name available `/gsd ...` or `scripts/gsd prompt ...` preflight commands and required Go/design skills from `required-skills-routing.md`
- owning PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE when the registry lacks `programming-loop`, without inventing that command
- receiving worker handoffs
- deciding whether a sub-PR can merge into the parent branch
- dispatching fresh-context read-only local Codex review against exact base/head identities
- dispositioning every actionable finding and re-running verification/re-review after head changes
- running independent Shepherd trajectory validation after clean review and before integration
- declaring final parent PR readiness for human approval

The orchestrator must not stop after describing next steps when all required inputs are available,
no human gate is triggered, and at least one sub-issue is worker-ready.

The meaning of `spawn` is runtime-generic:

- Claude Code: create Agent/Task workers.
- Codex: explicitly invoke Codex subagent tools or a custom Codex worker after assigning a separate
  git worktree or working directory for any worker that can edit files.
- OpenCode: invoke configured worker subagents or worker commands with `subtask: true`; keep the
  primary orchestrator in the main context.
- Future runtimes: create an isolated worker context with one issue, one branch, one write scope,
  one working directory, and the worker handoff template.

If the runtime has no worker mechanism and the requested mode requires agents, record
`not_spawned_runtime_capability_missing`.

If the runtime can spawn workers but cannot isolate mutating workers from the coordinator checkout,
record `not_spawned_isolation_missing`. Do not spawn code-writing workers into the same working tree
as the orchestrator. Read-only explorer/reviewer agents may still run in the shared checkout when
their prompt forbids edits.

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
- `sub_pr_reviewed`: exact-head local Codex review is clean and the independent Shepherd review
  transition received `PROCEED`.
- `provisionally_integrated`: sub-PR merged into the parent branch, but exact integrated-range
  verification/review/trajectory evidence is still pending.
- `parent_review_pending`: exact-head parent verification, local Codex review, or Shepherd
  validation is running for an integrated batch.
- `parent_review_clean`: the integrated batch has no unresolved actionable local Codex findings and
  its Shepherd validation passed.
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
- each mutating worker has a separate worktree or working directory
- shared parent artifacts are orchestrator-owned
- human gates are not crossed

When file ownership is unclear, run the sub-issues sequentially or split them further.

At every parent orchestration turn, record whether workers were spawned. If no worker is spawned
while work remains, the orchestrator must either take the local critical-path action that unblocks
workers or record exactly one blocker category:

- `not_spawned_dependency_blocked`
- `not_spawned_write_scope_collision`
- `not_spawned_human_gate`
- `not_spawned_isolation_missing`
- `not_spawned_runtime_capability_missing`
- `not_spawned_review_blocked`
- `not_spawned_verification_blocked`

Compact handoffs are allowed only for agent prose. Do not compact exact code, commands, test output,
review findings, security warnings, destructive-action warnings, ordered safety gates, or approval
gates in worker prompts or handoffs.

## Merge Policy

A sub-PR may merge into the parent branch only when:

- it targets the parent branch
- it uses `Refs #<sub-issue>` and `Refs #<parent-issue>`, not closing keywords
- targeted and issue-level verification pass
- CI checks pass or an infrastructure blocker is recorded
- exact-head local Codex findings are resolved and independent Shepherd validation passed
- the diff is within the sub-issue scope
- no requested-changes review is open
- no human gate is triggered

Review is bound to commit identity, not PR base behavior. A stacked sub-PR is review-complete only
when fresh-context local Codex review is clean for its exact base/head range and independent
Shepherd validation passes. Any head change invalidates both results and requires verification,
re-review, and revalidation.

The parent PR into `main` always requires human approval.

## Exact-Head Review And Trajectory Record

For every sub-issue, record:

- sub-issue number
- sub-PR URL
- parent PR URL
- exact base branch and SHA
- exact head branch and SHA
- reviewed commit range
- primary route: `local_codex`, `human`, or `blocked`
- review status: `pending`, `clean`, `comments_addressed`, or `blocked`
- fresh-context reviewer identity and findings/disposition artifact
- Shepherd status, verdict, trajectory score, and evidence artifact
- CI status and remaining human gates

Claude and GitHub Copilot are not required, requested, or fallback coverage for the canonical PM
route. Preserve legacy values only when reading truthful historical records.

## Output Requirements

The orchestrator must leave a reviewer able to answer:

- which sub-issues were launched
- which worker owned each sub-issue
- which branches and PRs were used
- which checks ran
- which exact-head local Codex and Shepherd evidence covered each sub-issue
- why any sub-issue was deferred or blocked
- whether the parent PR is ready for human approval
