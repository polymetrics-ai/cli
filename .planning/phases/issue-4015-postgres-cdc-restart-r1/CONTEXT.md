# PostgreSQL CDC Restart Recovery — Context

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `fm/cli-cdc-resume-fix-r1` → `integration/4015-mvp-flat-r1` → `main`
- Delivery: Direct pull request open against `integration/4015-mvp-flat-r1`, targeting release 0.2.1, with red/green live recovery evidence and the API-reported base verified.
- Working branch: `fm/cli-cdc-resume-fix-r1`
- Task: Diagnose and repair PostgreSQL CDC restart recovery without touching the 0.2.0 release branch or PR #4250. First reproduce the current failure. Then prove the accepted checkpoint resumes at the correct durable PostgreSQL position, with independently queried target rows arriving exactly once after interruption.
- Verification: Preserve the live failing reproduction; add a focused failing restart regression before production edits; run the PostgreSQL package and database-integration tests; independently query the target for exact before/interruption/after counts and key multiplicity; run the repository's generated, connector, build, vet, and documentation gates; inspect the CDC capability artifact honestly; verify the opened PR base through the GitHub API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Current restart failure is diagnosed from the running system | live | The existing PostgreSQL pipeline is interrupted and restarted; the red trace records the returned recovery outcome and whether the post-interruption row is absent from the independently queried target. |
| Restart resumes from the correct durable position | live | The restarted sync continues from the last durable checkpoint and delivers a row committed after interruption without re-bootstrap. |
| No loss or duplication | live | An independent SQL query records exact target counts before interruption, at interruption, and after restart, and groups the post-interruption key to prove multiplicity exactly one. |
| Regression is TDD-locked | live | A focused test fails before the production change and passes afterward while asserting delivered records/checkpoint position rather than exit status alone. |
| CDC certification evidence remains honest | live | The current `postgres_cdc_r1-capability-cdc.json` claim is compared with the observed implementation and updated if its certified truth changes. |

Every live check asserts destination, checkpoint, acknowledgement, or cleanup state; exit status alone is not evidence.

## Locked decisions and constraints

- The connector canon controls: PostgreSQL CDC is PostgreSQL 14+ streamed `pgoutput` v2 with bounded durable transaction staging and receipt before checkpoint/acknowledgement. Timestamp/cursor polling is not an acceptable substitute for CDC.
- The diagnosed mechanism mismatch is a hypothesis until the live red reproduction and checkpoint trace confirm or refute it.
- Validation will not be weakened merely to accept an incompatible checkpoint. Any accepted checkpoint must be decoded, bound to the same source/slot/publication/schema identity, and proven to resume at the right durable LSN.
- Scope is exactly the PostgreSQL connector and its issue evidence. A shared-runtime or unrelated connector change would require a separate foundation issue.
- Do not start, stop, restart, or update Colima or Docker. Reuse only the already-running endpoint exposed by the existing live harness.
- Do not touch `release/0.2.0-mvp`, PR #4250, or merge any PR.
- No new dependency, secret exposure, generic SQL surface, or direct connector hop.

## Required skills and references

- `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.
- `docs/connector-canon/INDEX.md`, `docs/connector-canon/IMPLEMENTATION-PROCEDURE.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, and `docs/architecture/connector-architecture-v2-design.md`.

## Lifecycle execution

The repo-local shell adapter resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`. Execution is inline/manual because the canonical single-worker contract forbids role spawning. This fallback does not weaken TDD, live evidence, verification, or review gates.
