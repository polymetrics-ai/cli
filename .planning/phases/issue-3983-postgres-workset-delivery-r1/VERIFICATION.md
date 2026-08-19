# Verification — Issue 3983

## Completed execution

| Acceptance criterion | Result | Observable evidence |
| --- | --- | --- |
| Managed keyed upserts from an immutable workset | Pass | `TestPostgresManagedTargetWorksetDeliveryLive` queries real PostgreSQL rows after inserts and a composite-key update. |
| Absence retains a row; explicit tombstone deletes it | Pass | The live test retains `retain/9` after its physical absence, then queries that it is gone only after a sealed explicit tombstone. |
| Stale approval is pre-mutation | Pass | `TestChangeDeliveryExecutorRefusesStaleWorksetApprovalBeforeSessionMutation` asserts fake begin/batch/commit/rollback, ledger, and baseline counters are all zero. |
| Receipt/unknown/failure baseline semantics | Pass | Unit tests retain the prior fake baseline on batch failure, ledger receipt failure, and unknown commit; successful delivery persists and re-reads the candidate baseline only after the receipt. |
| Destination ownership isolation | Pass | Concurrent distinct workspace/connection destinations retain their own file-store baseline entries; plan sealing rejects a mismatched destination. |
| Built runtime live proof | Pass | Required PostgreSQL `databaseintegration` command passed; trace records the direct Colima Docker invocation. |

## Commands and review

- `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/... ./internal/warehouse/... -count=1` — pass.
- `go test -race -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test(ChangeDelivery|FileChangeDelivery|PostgresManagedTargetWorksetDelivery)' -count=1` — pass.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`, `golangci-lint run ./internal/connectors/database/...`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`, and `git diff --check` — pass.
- The required tagged PostgreSQL dbtest is rerun after the final test-only path cleanup. See `traces/postgres-workset-delivery-live.txt`.
- Manual inline `code-review` checked approval ordering, explicit-delete-only input, receipt-before-baseline promotion, target/owner/OID/schema/key bindings, file-store atomicity, context handling, bounded artifacts, and absence of raw credentials or SQL. No findings remain; see `REVIEW.md`.

## Scope and parity

No command syntax, generated connector bundle, help/man page, or website surface is planned to
change. If implementation tracing proves otherwise, this checklist will be revised before code is
merged.
