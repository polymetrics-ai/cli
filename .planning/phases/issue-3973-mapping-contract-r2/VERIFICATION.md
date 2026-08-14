# Verification checklist — Issue #3973 mapping contract completion

## Behavioral proof

- [x] A V1 mapping exposes copied ordered target columns, source fields, nullability, and exact/lossless logical type plans through a sealed write plan.
- [x] A value round trips over an allowed widening mapping; an unrepresentable type/value is refused before a target projection.
- [x] The plan/approval equality includes mapping and tombstone counts.
- [x] An absent record does not delete a seeded fake row; a validated explicit tombstone does.
- [x] A missing/malformed/mismatched tombstone collection opens no session and mutates no fake target state.
- [x] `DeliveryReceiptV1` comes from a confirmed session commit and remains distinct from the ledger that makes it acknowledgement-eligible.

## Required local commands

- [x] `go test -timeout 20m ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/...`
- [x] `go test -race -timeout 20m ./internal/connectors/database -run 'Test(MappingContractV1|DatabaseWriteExecutor.*Tombstone)' -count=1`
- [x] `gofmt -w` for changed Go files
- [x] `go vet ./internal/connectors/database/... ./internal/synccontract/... ./internal/synctransport/...`
- [x] Scoped `make verify` gates listed in `AGENTS.md`
- [x] Inline `verify-work` and `code-review`, with no gap evidence required

## Deliberately not applicable

No PostgreSQL driver/DDL/SQL/live database test, CLI/help/manual/website change,
credential, generic write tool, capability promotion, source checkpoint, or
physical-absence deletion behavior is introduced in this driver-neutral
foundation.
