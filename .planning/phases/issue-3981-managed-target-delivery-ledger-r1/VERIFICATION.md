# Verification checklist — Issue 3981: durable target delivery ledger

## Behavioral proof

- [ ] Ledger identity binds asserted owner, target database, and immutable
      managed-target relation/StreamID only.
- [ ] A renamed source artifact with the same persisted StreamID resolves the
      previously recorded target delivery record.
- [ ] A fresh ledger instance on the same durable-store port resolves that
      same record after restart.
- [ ] Sibling relations in one owner namespace retain independent records; an
      update to one does not alter the other.
- [ ] Invalid/mismatched typed identity is refused before fake-store mutation.
- [ ] No transaction spool, source checkpoint, driver/DDL/SQL/mode/CLI code is
      introduced or repurposed.

## Local commands

- [ ] `go test -timeout 20m ./internal/connectors/database/... ./internal/app/...`
- [ ] `go test -race -timeout 20m ./internal/connectors/database -run 'TestManagedTargetDeliveryLedger' -count=1`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] Inline `verify-work` and `code-review`, including any GSD gap evidence.

## Deliberate not applicable

No command, flag, help text, manual, website page, connector definition,
PostgreSQL driver, DDL, SQL, credential, live target connection, reverse-ETL
execution, or source checkpoint changes. CLI/help/docs/website parity and live
driver proof are not applicable to this driver-neutral foundation slice; native
durability proof remains with the later driver issue.
