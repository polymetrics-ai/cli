# Plan — issue #117 Monday sensitive/admin policy

## Objective

Lock down Monday mutation/admin/sensitive operations as blocked metadata with explicit approval, typed confirmation, redaction, and documentation.

## GSD mode

- `scripts/gsd prompt plan-phase issue-117-monday-sensitive-admin --skip-research` generated this prompt.
- `programming-loop` command unavailable; manual TDD fallback active.

## Slice

1. Red test: verify Monday mutation operations and planned reverse-ETL commands have sensitive/admin/destructive policy metadata and docs name the policy.
2. Green: add/adjust docs or metadata only; do not enable writes.
3. Refactor: validate no `reverse_etl` command is implemented and no generic/raw write exists.

## Safety

No live writes, no secrets, no new dependencies. Every mutation stays blocked by default until plan → preview → approval → execute and typed confirmation policy is implemented.

## Verification

```bash
go test ./cmd/connectorgen -run 'TestMondaySensitiveAdminPolicy' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```
