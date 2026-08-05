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
- required skills or task type, including the installed GSD phase lifecycle through the repo-local
  adapter for implementation or behavior-changing work and required Go/design skills from
  `.agents/agentic-delivery/references/required-skills-routing.md`
- TDD plan
- verification commands
- safety notes
- source links
- automated review route expectations
- for CLI feature work, CLI help/manual/website parity expectations from `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`

If any of these are missing and the task is not trivial, update the issue or create a planning PR
before implementation.

## Required workflow

1. Read the issue first. Treat it as the task contract.
2. Read repo rules and required context named in the issue.
3. Confirm the issue maps to one primary PR. Split the issue if it is too large.
4. Load the skills required by `task-skill-matrix.yaml` for the issue task type and
   `.agents/agentic-delivery/references/required-skills-routing.md`. For Go work, start with
   `golang-how-to`; for CLI work include `golang-cli`; for website/design work include the relevant
   design skills such as `frontend-design`, `web-design-guidelines`, and
   `vercel-react-best-practices`.
5. For implementation or behavior-changing work, read
   `.agents/agentic-delivery/references/gsd-pi-adapter.md`, resolve every command through
   `scripts/gsd sources`, then follow `discuss-phase` → `plan-phase --tdd` → `execute-phase` →
   `verify-work`, including `plan-phase --gaps` and `execute-phase --gaps-only` when needed, then
   `code-review`. The pinned adapter does not provide `programming-loop`; never make it a required
   command. When compatible isolated runtime agents are unavailable or the canonical contract
   forbids role spawning, execute the generated prompts inline and record that fallback without
   weakening TDD evidence.
6. Create or update the GSD plan, TDD ledger, and verification checklist for the issue before
   production edits. The plan must name the slice boundaries, expected red/green/refactor evidence,
   verification commands, and commit/push checkpoints. For CLI command, flag, output, connector
   surface, or help-topic changes, include the CLI help/manual/website parity checklist from
   `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
7. Start a branch that includes the issue number when practical.
   - If the issue is a sub-issue in a parent roadmap, branch from the parent branch.
   - If the issue is a parent issue, branch from `main` and keep the parent PR human-gated.
   - If the issue is a sub-issue, confirm the parent PR from the parent branch to the default
     branch exists before treating the sub-issue as executable. Create a draft parent PR when it is
     missing and no human gate blocks creation. When the parent branch has no diff yet, create a
     deliberate parent seed commit first; use an empty commit only when a real scaffold file would be
     noise.
8. For behavior changes, write or update a failing test before production code.
9. For connector implementation lanes, confirm exactly one target connector before editing. If the
   issue needs shared runtime/tooling, schema, generated-index, or unrelated connector changes,
   stop the connector lane and split that work into a separate foundation issue/PR instead of
   absorbing it into the connector PR. no-mistakes validation must treat this as a stop/ask-user
   foundation split, not an auto-fixable connector diff.
10. Implement the smallest slice that satisfies the issue.
11. Run targeted tests, then broader verification from the issue. For CLI feature work, verify
    runtime help (`pm help <topic>`, `pm <namespace>`, `pm <command> --help`), docs under
    `docs/cli/**`, website docs under `website/**`, and generated help/manual artifacts as
    applicable.
12. Commit after each coherent green slice. Good checkpoints are plan-only, red-test, green
    implementation, refactor, and review-fix batches. Do not commit unrelated files.
13. Push each committed checkpoint to the active issue/PR branch after the relevant green gates so
    CI and automatic review can run regularly. Never push to `main`; stop only when a human gate is
    triggered.
14. Update phase or research artifacts when the issue asks for durable memory.
15. Open a PR with a Conventional Commit title and an issue-first body.
    - Use `Closes #N` or `Refs #N` when the PR is issue-backed.
    - Use `Refs #N` for sub-PRs that target a parent branch.
    - Use `Closes #N` only for PRs that target the default branch and complete the issue.
    - A no-mistakes-generated PR may satisfy the guard with its complete delivery record instead of
      a manual issue link only when the body has `## Intent`, `## What Changed`, `## Testing`, and
      `## Pipeline` sections plus the generated `git push no-mistakes` marker.
16. After implementation and local verification, choose the automated review route using
    `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`, then run the Claude
    review loop in
    `.agents/agentic-delivery/workflows/claude-review-loop.md`.
17. Confirm that Claude actually produced review records or that the stacked-PR parent-review
    fallback covers the sub-issue. A skipped-review status, rate-limit notice, or processing-only
    comment is not approval.
18. If Claude is rate-limited, skipped, disabled, paused, or unavailable and review coverage is
    blocking progress, request GitHub Copilot review once as a backup when enabled. Copilot
    comments are dispositioned like Claude comments, but Copilot review is not approval.
19. Reply to every actionable automated review item with the disposition template before resolving
    it.
20. Ensure accepted fix commits have been reviewed. Prefer Claude's automatic incremental review
    when active; request manual `@claude review` only when automatic review is paused,
    disabled, skipped, rate-limit retry is due, or the configured automatic pause threshold was
    reached.
21. Ping the human coordinator only after no actionable automated review findings remain or a
    recorded human review blocker remains.

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
to a parent roadmap. A sub-PR may advance to the canonical `integrate_sub_pr` state without human
approval only after all automated checks pass, automated review comments are resolved, review
coverage exists through the sub-PR, main-targeted parent PR, or an approved fallback route, and no
human gate is triggered.

For parent jobs, use `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`. The
compatibility filename now defines single-worker parent ownership: the canonical worker owns shared
parent artifacts, processes ready sub-issues without spawning roles, and makes sub-PR integration
decisions after the preserved check/review gates pass. GitHub issues, branches, PRs, and GSD
artifacts are the durable handoff state.

The parent PR into `main` always requires human approval.

## Output requirements

Every implementation PR must include:

- issue link, or the complete generated no-mistakes delivery record when the PR is not issue-backed
- connector implementation evidence when applicable: exactly one target connector, ownership guard evidence, changed-path compliance, and any foundation issue/PR path
- summary of changes
- red/green/refactor evidence when behavior changed
- GSD lifecycle evidence, including the `/gsd-*` or `scripts/gsd prompt ...` commands used and any
  explicit inline/manual fallback note
- Required Go/design skills loaded, with task-specific notes from
  `.agents/agentic-delivery/references/required-skills-routing.md`
- CLI help/manual/website parity evidence for CLI feature work, including bare namespace behavior
  such as `pm connectors` when relevant
- commit/push checkpoint summary
- verification commands and results
- safety notes for auth, secrets, writes, or data movement
- follow-up issues for work intentionally deferred
- automated review disposition summary, including accepted, declined, deferred, and human-gated
  findings, plus the Claude and Copilot route status
