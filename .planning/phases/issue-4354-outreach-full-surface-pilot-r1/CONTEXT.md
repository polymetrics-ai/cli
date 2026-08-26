# Outreach full-surface pilot context

## Task Delivery Header

- Issue: Closes #4354 — feat(connectors): make Outreach full-surface pilot auditable
- Base branch: main
- Merges into: main
- Delivery: A standalone PR against `main` is open, has the documented single-connector evidence, and is ready for independent audit; it is not merged by this task.
- Working branch: feat/4354-outreach-full-surface-pilot
- Task: Reconstruct and validate only the Outreach connector's source-backed surface from Batch 6–7 candidate `18248d233e6abd9d7ec03075a225cf35ee2f5399`, prove the usable command boundary without provider I/O, and disclose any active shared-foundation dependency.
- Verification: Source-lock and declaration audits; focused red/green tests; `go build ./cmd/pm`; generated evidence/certification/surface checks; representative fixture-only preflight commands; `git diff --check`; independent clean-worktree run; GitHub API PR-base readback.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every authoritative Outreach operation has a durable mapping or disposition | live | Source-lock and disposition inventories agree on exact operation identity/count; a missing row fails the focused audit. |
| Every lane has an explicit supported or not-applicable statement | live | The checked connector evidence counts all six fixed lanes and requires an evidence-bearing zero for absent lanes. |
| Valid user commands reach credential preflight | fake | Fixture-only transport and deliberately absent credential prevent real provider I/O; each test asserts the concrete `missing --credential` boundary rather than a successful exit. |
| Source identity/method/path substitution is refused | fake | A local test fixture changes the declared binding and asserts preflight rejection before a request; real mutation would require provider access and is prohibited. |
| Delete path is safely reachable but not executed | fake | Fixture-only command intentionally stops at credential preflight; running a destructive provider request is prohibited without plan, preview, approval, and credentials. |

## Scope decisions

- Target connector: **Outreach only**. Changed connector paths must remain under `internal/connectors/defs/outreach/`.
- Candidate artifacts are evidence to verify, not authority by themselves. Their asserted 259 operations and 96/163 ETL/write split remain provisional until source-lock and declaration reconciliation passes.
- PR #4350 owns the shared schema-v3 source-evidence reader, citation rendering, and wrong-source classification. This issue may not change shared readers, validators, generators, or schemas. A block remains recorded as a foundation dependency; rows must not be removed or relabelled to hide it.
- No credentials, provider network calls, reverse-ETL execution, source substitution, or merge are authorized.
- This is a connector surface/artifact import. Existing generic connector CLI behavior, help/manual/website pages, and namespace behavior are expected to be unchanged; parity will be explicitly checked and marked not applicable if confirmed.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `golang-lint` were loaded before implementation.

## Lifecycle execution

Resolved with `scripts/gsd doctor`, `scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}`, and `scripts/gsd prompt` for each lifecycle step. The Pi runtime is unavailable to this agent and the canonical single-worker contract forbids spawning GSD roles, so all five phases are executed inline with durable artifacts in this directory.
