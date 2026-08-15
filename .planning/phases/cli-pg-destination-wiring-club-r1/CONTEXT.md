# Context — PostgreSQL production transport wiring club

## Task Delivery Header

- Issues: Refs #3982, Refs #3983, Refs #3979
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Working branch: `fm/cli-pg-destination-wiring-club-r1`
- Delivery: Direct pull request against the integration base, followed by an API read-back of the observed base ref; no merge and no `no-mistakes` invocation.
- Task: Register PostgreSQL's managed destination, connect sealed warehouse worksets to its write driver, and dispatch PostgreSQL gap-free bootstrap through `app.Open` and the shipped command path.
- Verification: focused unit/race tests, a real Docker/Colima PostgreSQL source and target driven through production App/CLI composition, a built `pm`, generated-artifact drift checks, and scoped repository gates.

The captain explicitly clubs the three issues because they share the one definition-owned transport registration site introduced by #4093/PR #4156. This is the authorized exception to the ordinary one-primary-issue rule; the PR remains limited to these three audited production reachability residuals.

## Locked decisions

- `app.Open` remains the composition root. PostgreSQL publishes an exact destination declaration and connector-local factory; App does not select a provider by name or capability flag.
- Every route remains source → connection-owned WAL/Parquet + manifest → reopened sealed workset → PostgreSQL managed target. No direct connector pair, generic SQL, caller-selected relation, or arbitrary existing table is introduced.
- Public `metadata.json` capability `write` remains false. `change_capture` is not a destination mode, and unsupported target modes remain typed refusals rather than advertised successes.
- The PostgreSQL source declaration must admit real dynamically discovered relation names. The shipped executor resolves a relation from the connection/catalog contract, not from a caller-authored SQL string.
- Incremental PostgreSQL runs use `BootstrapCDC` when no durable bootstrap checkpoint exists and `ReadCDC` only when one does. Native source identities remain sealed in the provider checkpoint while App persists its connection-owned resume identity.
- PostgreSQL destination execution derives managed-target ownership from persisted workspace/source/connection/stream IDs, consumes the stage-owned Parquet artifact, and uses `ChangeDeliveryWorkset`/`ChangeDeliveryExecutor` for `incremental_upsert`.
- Approval remains plan → preview → single-use approval → execute. No destination DDL, control record, batch, receipt, baseline, or checkpoint can advance on a missing, stale, expired, or replayed approval.
- Existing #4125 and #4158 failures are not modified.

## Evidence strategy

The live proof launches independent PostgreSQL source and target systems with `internal/connectors/native/dbtest`, creates a project and credentials without exposing secret values, invokes production App/CLI composition (including a built binary), and queries source, target, private control/ledger rows, connection warehouse Parquet/manifests, persisted checkpoints, and slot state. A second opt-in proof reads a real GitHub issue with a real credential added through `pm credentials add`, then queries the live PostgreSQL row, receipt, artifacts, and checkpoint. An error or exit code alone never counts as proof.

The captain-required edge matrix is part of the TDD ledger and final PR body. Deterministic refusal/failure injection may use narrow fakes only to prove a pre-side-effect boundary; successful PostgreSQL behavior is always observed on the real databases.

## Required skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-database`, and `golang-documentation`.

The adapter doctor and agent-contract check passed, and sources for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved. The canonical issue worker contract forbids lifecycle-role spawning, so the official prompts are executed inline and these artifacts are the documented manual fallback.
