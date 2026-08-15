# Context — Issue 4094: PostgreSQL incremental dedupe history

## Task delivery header

- Issue: `Closes #4094 — feat(postgres): implement the incremental_dedupe_history managed target`
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 -> main`
- Delivery: direct PR against `integration/4015-mvp-flat-r1` with green checks
- Working branch: `fm/cli-4094-history-target-r1`
- Target connector: PostgreSQL only
- Required verification: `go test ./internal/connectors/native/postgres/... ./internal/connectors/database/...`; tagged PostgreSQL dbtest; CI

The named task-delivery-header template is absent from this integration base.
This context records its fields verbatim as the documented manual fallback.

## Binding decisions

- `incremental_dedupe_history` is admitted only for PostgreSQL source to
  PostgreSQL managed target. Every other source/destination route must return
  a route-specific typed refusal before a provider or database operation.
- Reuse `database.MappingContractV1`, `database.DeliveryReceiptV1`, the
  managed PostgreSQL write driver, sealed workset delivery, and the managed
  target delivery ledger. No private mapping or receipt contract is permitted.
- A successful route must persist the requested history semantics: keyed
  versions, validity-window close, soft-delete close, deterministic late/replay
  behavior, receipts, and restart recovery.
- This PR is limited to PostgreSQL and database shared contracts already owned
  by the named foundations. No work for #4125, #4136, or #4090 is absorbed.

## Foundation confirmation

Confirmed after rebasing to `origin/integration/4015-mvp-flat-r1`:

- `internal/connectors/database/mapping_contract.go`
- `internal/connectors/database/managed_target_delivery_ledger.go`
- `internal/connectors/native/postgres/managed_target_driver.go`
- `internal/connectors/database/workset_delivery.go`

## GSD execution

The generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` prompts were resolved through `scripts/gsd`.
This non-interactive worker has no compatible Pi isolated-agent runtime, so the
required lifecycle is executed inline and recorded in this phase directory.

## Required skills loaded

- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-database`
