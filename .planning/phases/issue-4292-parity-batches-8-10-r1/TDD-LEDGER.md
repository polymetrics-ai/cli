# Issue #4292 — TDD ledger

## Red

- Pending: the map-integrity assertion must fail before each batch because the
  ten target ledgers and their source locks/crosswalks do not exist.
- Pending: the same assertion must fail if a source ID is omitted, duplicated,
  placed in more than one primary class, or if a reverse-ETL row omits the
  locked `generic-typed-destination-executor` foundation gap.

## Green

- Pending: create source-locked ledgers for batch 8, execute the integrity
  assertion, connector validation, and surface-sync check, then commit.
- Pending: repeat for batch 9 and batch 10.

## Refactor / review

- Pending: inspect each final JSON ledger for false engine-gap claims,
  invented contracts, source location drift, delete coverage, and transport
  correction compliance.

## Fixed constraints

- Reverse ETL is a foundation gap in this task, not declaration-pending.
  Evidence: `internal/app/issue_label_warehouse_transport.go:85-95` at
  `acb85dc03`. No `transport_binding` action may be created.
- ETL is only declarable where the definition-owned source contract is actually
  satisfied.
