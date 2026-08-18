## Intent

Closes #3980

Derive immutable, StreamID-keyed Parquet delivery worksets for the PostgreSQL
parity foundation.

## What Changed

- Added `database.ChangeDeliveryWorkset`, keyed exclusively by the asserted
  `ManagedTargetDeliveryLedgerKey` identity from #3981 rather than mutable
  artifact/table/display text.
- Materializes an immutable complete projection, keyed insert/update delta,
  explicit tombstone stream, and unpromoted candidate baseline with a sealed
  manifest and content hash.
- Added real DuckDB/Parquet regression tests for composite-key delta behavior,
  null/type differences, explicit-only tombstones, source mutation immunity,
  zero-byte warehouse tables, bounded artifacts, cancellation cleanup, and
  corrupt-workset refusal.

## Safety and Scope

- Requires a finite `MaxArtifactBytes` ceiling and cleans failed/canceled
  staging artifacts. No credentials, target connection, target DML, checkpoint,
  or baseline promotion is performed.
- #3983/#3973 remain responsible for receipt-bound target delivery and baseline
  promotion. No CLI, connector definition, documentation surface, or capability
  claim changed.

## GSD / TDD Evidence

- Inline/manual lifecycle completed: `discuss-phase` → `plan-phase --tdd` →
  `execute-phase` → `verify-work` → `code-review`. The canonical single-worker
  contract forbids role spawning; the fallback is documented in the phase
  artifacts.
- Red evidence: missing immutable-workset API, missing artifact bound, and
  zero-byte warehouse Parquet failure. Green evidence and acceptance matrix:
  `.planning/phases/issue-3980-parquet-delivery-worksets-r1/`.
- Skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-database`, `golang-lint`.

## Testing

- `go test -timeout 20m ./internal/warehouse/... ./internal/connectors/database/... ./internal/synctransport/... -count=1`
- `go test -race -timeout 20m ./internal/connectors/database -run 'TestDeriveChangeDeliveryWorkset' -count=1`
- `go vet ./...`
- `go build ./cmd/pm`
- `golangci-lint run ./internal/connectors/database/...`
- Individual `make` gates: `tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, `release-workflow-check`.

## Review

- Manual standard-depth inline review: clean after fixing the mandatory finite
  artifact limit and zero-byte warehouse-Parquet handling.
- Claude automated review is expected on PR open; no Copilot fallback requested.
