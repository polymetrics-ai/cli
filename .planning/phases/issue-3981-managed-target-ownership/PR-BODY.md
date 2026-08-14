## Intent

Refs #3981

Refs #3972

Implement the shared, driver-neutral managed-target ownership and provisioning
contract for the PostgreSQL parity parent. This child targets
`feat/3972-postgres-parity`, not `main`.

## What Changed

- Derive opaque physical namespace/relation names from the source warehouse
  artifact identity (workspace, source connector, source connection, table).
- Add typed owner, target ref, schema hash/version, native relation identity,
  control record, immutable provisioning plan, observation, and narrow driver
  port values.
- Enforce create-or-assert: create only when namespace and control are absent;
  otherwise admit only an exact owner/ref/native/schema match and fail closed.
- Refuse unreadable, missing, foreign, colliding, orphaned, moved/replaced, and
  schema-drifted state. There is no customer-table adoption or schema evolution.
- Require a driver-neutral target-scoped lock before observation through final
  reassertion. #4038 records the review-found cross-provisioner correction.

## TDD and GSD Evidence

- `TestManagedTargetProvisioningTruthTable` was RED before the shared contract
  existed; the preserved output is in
  `.planning/phases/issue-3981-managed-target-ownership/traces/`.
- GREEN covers absent namespace, first/repeat create, correct/missing/foreign/
  unreadable owner, collisions, native replacement, schema drift, orphaned
  control, invalid plan, cancellation, post-create reassertion, and concurrent
  calls across two provisioners.
- #4038 was created and linked from #3981 before the review correction; its
  separate RED/GREEN evidence is preserved alongside the original truth table.
- Required lifecycle commands were resolved through `scripts/gsd sources` and
  generated with `scripts/gsd prompt`. This issue-local child records the
  canonical explicit inline-GSD fallback because the contract permits one
  inline worker only.

## Testing

- `go test -race -timeout 20m -count=10 ./internal/connectors/database -run '^TestManagedTargetProvisioningTruthTable$'`
- `go test -timeout 20m -count=1 ./internal/connectors/database`
- `go test -timeout 20m -count=1 ./internal/warehouse`
- `go test -timeout 20m -count=1 ./internal/synccontract`
- `go test -timeout 20m -count=1 ./internal/app`
- `go test -timeout 20m -count=1 ./internal/cli`
- `go vet ./...`; `go build ./cmd/pm`; formatting/tidy/package+repo lint
- Individual docs/smoke/agent-contract/connectorgen/certification/runtime-
  preflight/canon/boundary/release gates (all green; boundary clean)

## Safety and Scope

- No PostgreSQL-specific driver/DDL, generic SQL, target executor, direct
  write/query surface, CDC, transport, credential access, or capability
  promotion is included.
- The fake driver is the only implementation used for the contract proof.
- CLI help/manual/website parity is not applicable: no CLI surface changed.
- The bounded #3995 Shepherd-compatible record is `RETRY` only for a future
  certification transition; it is not an approval or merge authorization.

## Pipeline and Review

- Branch checkpoints: `0d49238ea`, `fcb29c9e4`, `2df69e5d1`, `20d93ae9d`, and
  `1a6140a36`, all pushed.
- Inline code review: PASS; one review correction, #4038, already verified.
- no-mistakes: pending shared-run availability at PR-body preparation time.
- Claude automatic review: pending draft PR creation. Copilot is fallback-only
  if the Claude route is unavailable.

## Follow-up

Refs #4038 (cross-provisioner lock correction; implemented in this PR).

Human review and merge remain required. Do not merge this child into the parent
or `main` automatically.
