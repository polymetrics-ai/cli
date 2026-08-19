# Verification — Issue 3980: immutable Parquet delivery worksets

## Acceptance evidence

| Acceptance criterion | Result | Observable evidence |
| --- | --- | --- |
| Deterministic inserts, updates, unchanged rows, explicit tombstones, composite keys, null/type edges | Pass | Real DuckDB/Parquet fixture returns exactly three delta rows (update, insert, string-vs-number type change), excludes the unchanged composite-key row, and preserves the explicit tombstone. |
| Physical absence never emits a tombstone | Pass | Fixture baseline contains `gone/9`, source omits it, and the reopened tombstone list contains only the supplied `removed/4` event. |
| Identical inputs are immutable; schema/key/destination invalidate reuse | Pass | Two derivations compare byte-identical identities/content hashes. Destination, schema, and key mutations each change identity; mutable source-artifact rename does not. Replacing source Parquet leaves the first sealed projection/hash unchanged. |
| Baseline advancement requires later receipt | Pass for this foundation boundary | The supplied baseline's bytes are compared before/after success and every refusal. Derivation creates only a separate candidate; it has no target, receipt, ledger-write, or checkpoint port. |
| Resources are bounded and cleanup is safe | Pass | Required `MaxArtifactBytes` rejects over-bound input before root creation; staged cancellation leaves zero children; corrupt immutable artifact is refused rather than replaced; empty zero-byte warehouse Parquet yields zero records. |
| Focused live implementation tests with Red/Green evidence | Pass | Red compiler/behavior failures and green commands are stored in `traces/`; tests use the repository's embedded DuckDB Parquet writer/reader. |

## Commands

- `go test -timeout 20m ./internal/warehouse/... ./internal/connectors/database/... ./internal/synctransport/... -count=1` — pass.
- `go test -race -timeout 20m ./internal/connectors/database -run 'TestDeriveChangeDeliveryWorkset' -count=1` — pass.
- `go vet ./...` — pass.
- `go build ./cmd/pm` — pass.
- `golangci-lint run ./internal/connectors/database/...` — pass, 0 issues.
- Individual final-tree `make` gates (`tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, `release-workflow-check`) — pass.

## Not applicable

No command, help, manual, website, generated connector surface, credential, or
live target database changed, so CLI parity and credentialed integration checks
are not applicable.
