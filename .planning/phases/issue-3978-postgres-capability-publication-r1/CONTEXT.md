# Context — Issue #3978: final PostgreSQL certification and publication

## Task delivery header

- Issue: `Refs #3978 — Postgres Parity: certify parity and publish truthful capabilities`.
- Base branch: `integration/4015-mvp-flat-r1` at `4a0289bcc490d705b12640093f5779df48a28cfe`.
- Delivery: one direct PR, `fm/cli-3978-postgres-certify-publish-r1` to `integration/4015-mvp-flat-r1`.
- Target connector: PostgreSQL only. No generic SQL writer, direct source-to-target route, API CDC sink, broker, MCP, or UI behavior.

## Binding reconciliation

1. #3978's `incremental_dedupe_history` rejection clause is stale: current PostgreSQL declares and live-proves all six managed-target modes, including history.
2. #3978's Podman-only wording is stale: `native/dbtest` supports Docker and the evidence run used the explicit shared Colima Docker socket without restart or reconfiguration.
3. `sync_transport.json` already admits the truthful narrow publication: closed `postgres_managed_target`, six named modes, fixed strategies, warehouse receipt acknowledgement, and no caller-selected SQL/target.
4. `metadata.json.capabilities.write` and `connectors.Capabilities.Write` admit only a generic boolean. It cannot express the closed destination transport; `write=true` promises arbitrary generic writer semantics that PostgreSQL's direct `Connector.Write` intentionally refuses.
5. The fresh PostgreSQL 16 aggregate profile is sound proof that all six closed strategies executed with independent read-back. It cannot be bound or published as `capability:write`. The source-controlled twelve exact `sync_mode` records remain the publishable proof for the narrow destination contract.
6. `cdc=true` is separate: PostgreSQL 14+ pgoutput is proven only from database source into the connection-owned warehouse, after durable staging and receipt. `query=false` stays exact.

## GSD execution and skills

The project-local GSD adapter was checked with `scripts/gsd doctor`, sources were resolved for discuss/plan/execute/verify/review, and the gap prompts were generated inline because the canonical single-worker contract forbids lifecycle-role spawning. Loaded skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-lint`, and `golang-documentation`.
