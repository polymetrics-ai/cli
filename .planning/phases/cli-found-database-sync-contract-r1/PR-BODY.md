<!-- Issue-first body for the active pull request. -->

## Intent

Complete Issue #3810's shared database-sync foundation without advertising a provider capability.
The change establishes durable, opaque checkpoint semantics and delete-aware history for later
native engine lanes to consume.

## Linked Issue

Closes #3810

## What Changed

- Added `internal/synccontract` with the closed sync-mode vocabulary, versioned opaque checkpoints,
  typed recovery outcomes, acknowledgement-gated commitment, tombstone/history rules, and immutable
  conformance fixtures.
- Migrated ETL stream state to checkpoint envelopes, requiring explicit rebootstrap for legacy
  cursors and refusing unadmitted contract modes before source reads.
- Hardened ETL and reverse-ETL durability so complete destination acknowledgement and durable state
  persistence precede checkpoint advancement or irreversible approval consumption.

## Testing

- `go test -count=1 ./internal/synccontract`
- `go test -race -count=1 ./internal/state`
- `go test -count=1 ./internal/durability`
- `go test -count=1 ./internal/app`
- `go vet ./...`
- `go build ./cmd/pm`

## Delivery Evidence

The red/green ledger, manual GSD fallback, focused verification, and review record are retained in
`.planning/phases/cli-found-database-sync-contract-r1/`. No CLI, connector bundle, credential,
live-provider, generic SQL, generic HTTP-write, or shell surface was added.

This current-main successor to #3879 inherits the certification harness fix from #3878; it does
not vendor the unrelated certification change.

## Pipeline

`require-linked-issue` reads this live GitHub PR body. The `## Intent` and `Closes #3810` sections
are intentionally retained here so the guard remains verifiable from the PR metadata.
