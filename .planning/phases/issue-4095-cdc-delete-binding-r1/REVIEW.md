# Focused code review — Issue 4095

Manual inline review substitutes for the generated reviewer because this
direct, non-numeric phase cannot run compatible isolated GSD roles.

## Scope reviewed

- `internal/connectors/native/postgres/cdc_tombstone.go`
- `internal/connectors/database/mapping_contract.go`
- `internal/connectors/database/workset_delivery.go`
- the matching unit and tagged integration tests

## Checks

- The CDC adapter accepts only `delete`, requires a parseable source LSN,
  configured non-null source keys, and a non-zero transaction-local ordinal.
  It has no database, HTTP, shell, or credential access.
- Its digest includes LSN, canonical key JSON, and ordinal, so two deletes at
  one LSN cannot collapse into a duplicate envelope identity.
- `MapTombstone` allows only the configured source key set and reuses the
  sealed source→target names; an absent input record cannot invoke it.
- Existing workset execution now calls the same mapping method. The native
  history implementation remains the sole owner of validity-window close and
  no #4154 files or behavior were modified.
- Unit failures occur before a write input exists. Tagged tests assert actual
  target rows/validity columns, not successful return values.

## Result

PASS — no actionable correctness, security, lifecycle, or scope findings.
