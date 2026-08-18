Closes #3982

## Intent

Implement PostgreSQL managed-target provisioning and the private, session-bound
write driver using the shared typed mapping, write-session, tombstone, and
receipt contracts. Public PostgreSQL write capability remains `false` pending
#3978 certification.

## What changed

- Attached `MappingContractV1` to typed provisioning, then used it to create
  PostgreSQL private namespace/control tables and a closed, typed target DDL
  layout.
- Implemented target identity/owner/control reassertion, durability preflight,
  atomic full overwrite, append, keyed upsert, dedupe, typed tombstones, and
  shared delivery receipts/unknown-commit reporting.
- Added live dbtest evidence for the first-create path, the five permitted
  modes, explicit-only deletion, and every refusal preserving database state.

## Evidence and safety

| Criterion | Evidence |
| --- | --- |
| Target creation and identity | Live PostgreSQL test reads private owner/control rows, catalog OIDs, and mapped target columns after create and reassert. |
| Refusal safety | Foreign owner, collision, unreadable controls, OID replacement, schema drift, unsupported mapping, and durability failures re-read state and prove no mutation. |
| Five modes | Live rows prove `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, and `incremental_dedupe`. |
| Explicit-only deletion | A row absent from a normal batch remains; an explicit shared tombstone deletes its composite key. |
| Atomicity | Statement failure and cancellation retain the pre-overwrite data; a real connection close after apply produces one shared unknown-commit outcome. |
| Capability fence | Public connector metadata stays `write=false`; legacy `Connector.Write` remains unsupported. |

No credentials, DSNs, or raw SQL-write surface are introduced. PostgreSQL
identifiers are validated/quoted through the native closed helper; values are
bound parameters.

## GSD / TDD / review

- Inline/manual GSD fallback: resolved `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review` through `scripts/gsd
  prompt ...`; the canonical worker contract prohibits spawning lifecycle
  roles. Durable Red/Green evidence is in
  `.planning/phases/issue-3982-postgres-write-driver-r1/`.
- Required skills loaded: `golang-how-to`, `golang-database`,
  `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-context`, `golang-concurrency`,
  `golang-design-patterns`, and `golang-structs-interfaces`.
- Manual code review found and fixed declared-collation assertion drift and a
  nullable tombstone-key no-op. No unresolved finding remains. Claude is the
  automatic review route for this PR; no Copilot fallback has been requested.

## Verification

- `go test -timeout 20m -count=1 ./internal/connectors/native/postgres/... ./internal/connectors/database/... ./internal/connectors/engine`
- `go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database ./internal/connectors/engine`
- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres`
- `go vet ./internal/connectors/native/postgres/... ./internal/connectors/database/... ./internal/connectors/engine`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check`

All passed locally. No changes for #4125, #4136, or #4090 are included.
