# Verification checklist — Issue #3973 mapping contract completion

## Behavioral proof

- [ ] A V1 mapping exposes copied ordered target columns, source fields, nullability, and exact/lossless logical type plans through a sealed write plan.
- [ ] A value round trips over an allowed widening mapping; an unrepresentable type/value is refused before a target projection.
- [ ] The plan/approval equality includes mapping and tombstone counts.
- [ ] An absent record does not delete a seeded fake row; a validated explicit tombstone does.
- [ ] A missing/malformed/mismatched tombstone collection opens no session and mutates no fake target state.
- [ ] `DeliveryReceiptV1` comes from a confirmed session commit and remains distinct from the ledger that makes it acknowledgement-eligible.

## Required local commands

- [ ] `go test -timeout 20m ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/...`
- [ ] `go test -race -timeout 20m ./internal/connectors/database -run 'Test(MappingContractV1|DatabaseWriteExecutor.*Tombstone)' -count=1`
- [ ] `gofmt -w` for changed Go files
- [ ] `go vet ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/...`
- [ ] Scoped `make verify` gates listed in `AGENTS.md`
- [ ] Inline `verify-work` and `code-review`, with any gap evidence recorded

## Deliberately not applicable

No PostgreSQL driver/DDL/SQL/live database test, CLI/help/manual/website change,
credential, generic write tool, capability promotion, source checkpoint, or
physical-absence deletion behavior is introduced in this driver-neutral
foundation.

