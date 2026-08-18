# Verification checklist — Issue 3981: durable target delivery ledger

## Behavioral proof

- [x] Ledger identity binds asserted owner, target database, and immutable
      managed-target relation/StreamID only.
- [x] A renamed source artifact with the same persisted StreamID resolves the
      previously recorded target delivery record.
- [x] A fresh ledger instance on the same durable-store port resolves that
      same record after restart.
- [x] Sibling relations in one owner namespace retain independent records; an
      update to one does not alter the other.
- [x] Invalid/mismatched typed identity is refused before fake-store mutation.
- [x] No transaction spool, source checkpoint, driver/DDL/SQL/mode/CLI code is
      introduced or repurposed.

## Local commands

- [x] `go test -timeout 20m ./internal/connectors/database/... ./internal/app/...`
- [x] `go test -race -timeout 20m ./internal/connectors/database -run 'TestManagedTargetDeliveryLedger' -count=1`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`
- [x] Inline `verify-work` and `code-review`, including any GSD gap evidence.

## Deliberate not applicable

No command, flag, help text, manual, website page, connector definition,
PostgreSQL driver, DDL, SQL, credential, live target connection, reverse-ETL
execution, or source checkpoint changes. CLI/help/docs/website parity and live
driver proof are not applicable to this driver-neutral foundation slice; native
durability proof remains with the later driver issue.
