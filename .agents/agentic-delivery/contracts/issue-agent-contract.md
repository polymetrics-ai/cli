# Generic issue-to-PR agent contract

Use this contract when an issue is assigned to an implementation agent.

## Required input

The issue must provide:

- objective
- background
- scope
- non-goals or exclusions
- acceptance criteria
- required reading
- required skills or task type, including `gsd-programming-loop` for implementation or
  behavior-changing work
- TDD plan
- verification commands
- safety notes
- source links

If any of these are missing and the task is not trivial, update the issue or create a planning PR
before implementation.

## Required workflow

1. Read the issue first. Treat it as the task contract.
2. Read repo rules and required context named in the issue.
3. Confirm the issue maps to one primary PR. Split the issue if it is too large.
4. Load the skills required by `task-skill-matrix.yaml` for the issue task type.
5. For implementation or behavior-changing work, load and follow `gsd-programming-loop` before
   coding. If local GSD scripts are unavailable, run the manual GSD loop and record the fallback in
   the phase or PR artifacts; do not skip TDD evidence.
6. Create or update the GSD plan, TDD ledger, and verification checklist for the issue before
   production edits. The plan must name the slice boundaries, expected red/green/refactor evidence,
   verification commands, and commit/push checkpoints.
7. Start a branch that includes the issue number when practical.
   - If the issue is a sub-issue in a parent roadmap, branch from the parent branch.
   - If the issue is a parent issue, branch from `main` and keep the parent PR human-gated.
8. For behavior changes, write or update a failing test before production code.
9. Implement the smallest slice that satisfies the issue.
10. Run targeted tests, then broader verification from the issue.
11. Commit after each coherent green slice. Good checkpoints are plan-only, red-test, green
    implementation, refactor, and review-fix batches. Do not commit unrelated files.
12. Push each committed checkpoint to the active issue/PR branch after the relevant green gates so
    CI and automatic review can run regularly. Never push to `main`; stop only when a human gate is
    triggered.
13. Update phase or research artifacts when the issue asks for durable memory.
14. Open a PR with a Conventional Commit title and `Closes #N` or `Refs #N` in the body.
    - Use `Refs #N` for sub-PRs that target a parent branch.
    - Use `Closes #N` only for PRs that target the default branch and complete the issue.
15. After implementation and local verification, run the CodeRabbit review loop in
    `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`.
16. Reply to every actionable CodeRabbit item with the disposition template before resolving it.
17. Ensure accepted fix commits have been reviewed. Prefer CodeRabbit's automatic incremental review
    when active; request manual `@coderabbitai review` only when automatic review is paused,
    disabled, skipped, rate-limit retry is due, or the configured automatic pause threshold was
    reached.
18. Ping the human coordinator only after no actionable CodeRabbit findings remain.

## Hard stops

Stop and ask for human approval before:

- token scope changes or `gh auth refresh`
- reading, requesting, printing, storing, or inventing secrets
- new dependencies
- destructive external actions
- production deploys
- broad generated-file rewrites not named in the issue
- weakening tests or quality gates
- exposing generic shell, generic HTTP write, generic SQL write, or unrestricted raw API tools
- executing reverse ETL without plan, preview, approval, execute
- merging a parent PR into `main`

## Parent/subissue work

Use `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md` when the issue belongs
to a parent roadmap. A sub-PR may be merged into the parent branch without human approval only after
all automated checks pass, CodeRabbit comments are resolved, and no human gate is triggered. The
parent PR into `main` always requires human approval.

## Output requirements

Every implementation PR must include:

- issue link
- summary of changes
- red/green/refactor evidence when behavior changed
- GSD programming-loop evidence or an explicit manual-GSD fallback note
- commit/push checkpoint summary
- verification commands and results
- safety notes for auth, secrets, writes, or data movement
- follow-up issues for work intentionally deferred
- CodeRabbit disposition summary, including accepted, declined, deferred, and human-gated findings
