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
- required skills or task type
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
5. Start a branch that includes the issue number when practical.
   - If the issue is a sub-issue in a parent roadmap, branch from the parent branch.
   - If the issue is a parent issue, branch from `main` and keep the parent PR human-gated.
6. For behavior changes, write or update a failing test before production code.
7. Implement the smallest slice that satisfies the issue.
8. Run targeted tests, then broader verification from the issue.
9. Update phase or research artifacts when the issue asks for durable memory.
10. Open a PR with a Conventional Commit title and `Closes #N` or `Refs #N` in the body.
    - Use `Refs #N` for sub-PRs that target a parent branch.
    - Use `Closes #N` only for PRs that target the default branch and complete the issue.
11. After implementation and local verification, run the CodeRabbit review loop in
    `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`.
12. Reply to every actionable CodeRabbit item with the disposition template before resolving it.
13. Request incremental CodeRabbit review after accepted fixes, then ping the human coordinator only
    after no actionable CodeRabbit findings remain.

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
- verification commands and results
- safety notes for auth, secrets, writes, or data movement
- follow-up issues for work intentionally deferred
- CodeRabbit disposition summary, including accepted, declined, deferred, and human-gated findings
