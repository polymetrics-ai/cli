# Issue 4015 Cross-System Pipeline Certification — Context

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Direct pull request open against `integration/4015-mvp-flat-r1`, with live route evidence, local verification, and API-reported base verification.
- Working branch: `fm/cli-sync-e2e-crossystem-r1`
- Task: Certify PostgreSQL → warehouse → GitHub, GitHub → warehouse → PostgreSQL, GitHub → warehouse → GitHub, and job-backed ETL/reverse-ETL through a persisted flow and installed schedule. Each route must report a verdict, independent destination read-back, exact count and named sample, second-run behavior, and cleanup. Product defects are evidence only and will not be fixed here.
- Verification: Fresh `pm` binary commands; live GitHub reads against only `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`; real PostgreSQL container reads through pgx; durable warehouse, flow, schedule, and run-state inspection; independent GitHub fixture deletion followed by HTTP 404; targeted Go tests and repository gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| PostgreSQL → GitHub route | live | A uniquely named `pm-cert-` GitHub fixture changes only after the warehouse-mediated approved run; a separate GitHub client reads exact content and count. |
| GitHub → PostgreSQL route | live | A real GitHub record is materialized through Parquet and a separately opened pgx connection reads the exact managed-table count and named content. |
| GitHub → GitHub route | live | A real GitHub source record crosses the owned warehouse and an approved typed GitHub action changes a separately read destination fixture. |
| Incremental behavior per route | live | A second run records the observed skipped, duplicated, updated, or failed behavior and the destination is independently recounted. |
| Flow and schedule | live | The persisted flow runs the approved job through the exact installed scheduler payload; schedule inspection reports terminal state and the destination is independently read. |
| Cleanup | live | Every task-created GitHub object is removed and an independent request returns HTTP 404; task-owned database/container resources are absent after cleanup. |

## Assertion Rule

No exit code is accepted as route proof. Each passing route requires an independently observed destination state change that would be absent if the pipeline did nothing.

## Constraints

- Credentials are read inline from macOS Keychain service `pm-cert-classic` only at execution time and supplied to the test through environment/in-memory APIs. They never enter argv, files, fixtures, logs, planning artifacts, or PR text.
- No ambient GitHub CLI authentication is used.
- All GitHub writes are confined to `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` and uniquely prefixed `pm-cert-`.
- Reverse delivery remains plan → preview → stdin approval → execute.
- Existing Colima/Docker/Podman state may be inspected and reused but not started, stopped, or restarted.
- Product code fixes, CDC remediation, the 5 GB scale run, and writes outside the named fixture repository are out of scope.

## Foundation Check

| Need | Required proof in this phase | Missing-foundation handling |
| --- | --- | --- |
| Database source and managed target | Real PostgreSQL harness plus independent pgx read-back | Record `broken-with-evidence`; do not add generic SQL paths. |
| GitHub source and typed destination | Real GitHub API read/write using declared connector streams/actions | Record exact preflight/runtime failure; do not bypass with generic HTTP writes. |
| Warehouse mediation | Connection-owned WAL/Parquet and manifest/workset evidence | Route is not proven without the warehouse artifact. |
| Flow and schedule | Stored approved job, persisted flow, installed scheduler payload, terminal fire state | Record `broken-with-evidence`; do not treat direct one-shot success as schedule proof. |

## Required skills

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-database`, `golang-design-patterns`, and `golang-structs-interfaces`.

CLI help/docs/website parity is expected to be not applicable because this lane changes certification tests/evidence only. Any discovered command-contract change would require a separately scoped issue.

