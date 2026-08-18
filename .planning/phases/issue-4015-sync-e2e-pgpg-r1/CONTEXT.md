# Issue 4015 Sync Pipeline E2E — Context

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `fm/cli-sync-e2e-pgpg-r1` → `integration/4015-mvp-flat-r1` → `main`
- Delivery: Direct pull request open against `integration/4015-mvp-flat-r1`, with live control evidence, cleanup evidence, and the API-reported base verified.
- Working branch: `fm/cli-sync-e2e-pgpg-r1`
- Task: Certify PostgreSQL → PostgreSQL first through the real source → warehouse → managed destination pipeline, prove exact destination state independently, repeat the run to characterize `incremental_upsert`, and report all four requested route verdicts honestly within the time box. Do not fix product code.
- Verification: Run the opt-in real PostgreSQL binary harness against an already-running local container runtime, inspect its observable row/content assertions, run repository planning/document gates, verify fixture cleanup, and read the opened pull request base from the GitHub API through `gh-axi`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| PostgreSQL → PostgreSQL end-to-end result | live | A fresh source has 1,001 named rows; a separately opened PostgreSQL target connection must report exactly 1,001 rows and the named sample must equal the seeded source values. |
| Incremental behavior on a second run | live | A second approved `incremental_upsert` run must leave the target at 1,001 rows with unchanged content, proving replay/upsert rather than duplication. |
| Per-route verdicts | live | The final verification artifact records proven, broken-with-evidence, or not-attempted-and-why for all four requested routes. |
| Fixture cleanup | live | The database integration harness owns a unique container and volume; post-run runtime inspection must show both absent. GitHub residue is checked only if a GitHub route is attempted. |

Every live check asserts destination or cleanup state; exit status alone is not evidence.

## Locked decisions

- PostgreSQL → PostgreSQL is the control and must run before any GitHub route.
- The current branch adds certification evidence only. Product defects are findings, not repair scope.
- Credentials enter only through stdin or environment-backed credential ingestion and are never written to evidence.
- Reverse ETL follows plan → preview → approval → execute.
- Time box wins: after a decisive control result, unattempted GitHub routes are reported rather than delaying the control PR.

## Required skills and references

- `github-issue-first-delivery`
- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-database`
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`
- `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, and `docs/architecture/connector-architecture-v2-design.md`

## Lifecycle execution

The repo-local shell adapter generated `discuss-phase --auto` and `plan-phase --tdd --skip-research --auto` prompts. Execution is inline/manual because this task's canonical single-worker contract and supervisor brief forbid role spawning. This fallback does not weaken live assertions or review gates.
