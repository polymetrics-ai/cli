# Context: Issue #4283 — Batch-1 source-rigidity R2

## Task Delivery Header

- Issue: Refs #4283 — Batch-1 source-transport declarations.
- Base branch: `main` (GitHub API read-back for existing PR #4294: `base=main`, `head=fm/cli-top100-declaration-batch-r1`, `head_sha=6ec34964fda7a78a1736a5cd8933a46418346261`).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: Existing PR #4294 contains only scoped, reviewed, verified work and its exact remote head equals the freshly reviewed SHA; no merge is performed by this task.
- Working branch: `fm/cli-top100-declaration-batch-r1` (checked out detached at its fetched remote tip in this task worktree).
- Task: Reconcile the ten Batch-1 source-lock cohorts to the required source identity disposition invariant without credentialed or provider-live I/O, including only the shared foundations explicitly needed by the named cohort.
- Verification: GSD lifecycle evidence; immutable discovery/review ledger; red/green/refactor evidence for every authorized slice; per-connector source/mapping/reachability reconciliation; focused hermetic tests; applicable generator/CLI/docs checks; exact-SHA fresh review; PR check diagnosis; normal fast-forward push only after all gates are satisfied.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The task has an exact existing PR base and head. | live | `gh api /repos/polymetrics-ai/cli/pulls/4294` returned `base=main`, the required head ref, and SHA `6ec34964fda7a78a1736a5cd8933a46418346261`. |
| No production change occurs before a valid TDD plan and scope gate. | live | Initial `git status --short` was clean; this phase directory contains planning/review evidence only. |
| A provider request is not made. | live | The command ledger contains only local repository reads, generator/GSD discovery, and GitHub metadata reads; no connector credential or provider endpoint is used. |

## Immutable discovery frame

- Frozen remote source SHA: `6ec34964fda7a78a1736a5cd8933a46418346261`; it equals `origin/fm/cli-top100-declaration-batch-r1` after fetch and includes required ancestor `6ec34964fda7a78a1736a5cd8933a46418346261`.
- Existing PR: #4294, `feat(connectors): declare batch-one source transports`, API base `main`, head `fm/cli-top100-declaration-batch-r1`.
- Batch-1 source-operation census from the supplied immutable audit: Asana 249, Bitbucket 297, CircleCI 111, Docker Hub 54, GitLab 1,752, Jira 617, Notion 49, Sentry 223, Stripe 589, Vercel 400; total 4,341 REST source identities. A complete per-identity lane matrix remains a planned prerequisite, not evidence of usable execution.
- Source audit: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-source-lock-rigidity-audit-r1/report.md` identifies F1–F2, P1–P6, R1–R4, and E1–E3. Its documented source counts and explicit evidence gaps are retained as the pre-edit baseline.

## Scope-gate result

Inbox message `006.msg` (2026-08-29T13:45:36Z) is an explicit captain authorization to change the repository delivery policy itself: the obsolete one-connector/separate-foundation constraint is replaced by a bounded named-cohort policy. This is not a quality-control exception.

The authorized cohort is Asana, Bitbucket, CircleCI, Docker Hub, GitLab, Jira, Notion, Sentry, Stripe, and Vercel. Before production edits it must retain the immutable source-lock ledger and per-connector ownership/path matrix. A named shared foundation and its affected mappings may share this PR only after Foundation Atlas classification; source evidence, red-green-refactor, approval semantics, no provider I/O, review, CI, and captain-only merge remain mandatory. Unrelated connectors and unrelated runtime behavior remain excluded.

## GSD and skills

- `scripts/gsd doctor`, all required `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check` passed.
- Prompts generated and executed inline as planning guidance: `discuss-phase issue-4283-top100-declaration-batch-r1` and `plan-phase issue-4283-top100-declaration-batch-r1 --tdd`.
- Manual fallback: this issue is not an adapter-resolvable roadmap phase and the active task rules prohibit role spawning; generated GSD prompts are executed inline with evidence recorded here.
- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-graphql`, and repository-local `connector-lane-build-order`.
