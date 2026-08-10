# Phase 600 TDD ledger

**Issue:** #3984  
**Execution mode:** Inline single-worker fallback; no GSD role was spawned by
the canonical delivery contract.

## Capability matrix — Plan 01

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
|---|---|---|---|---|
| Source-derived inventory and executor status | Pending: add focused `cmd/connectorgen` tests before generator production code | Pending | Pending | planned |
| Strict cells and certification completeness | Pending: missing live evidence, stub write, and bad N/A tests | Pending | Pending | planned |
| Deterministic artifact/drift gate | Pending: mismatch test before writer implementation | Pending | Pending | planned |

## Pair-flow matrix — Plan 02

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
|---|---|---|---|---|
| Endpoint roles and pair identity | Pending: roles/reasons and source+destination key tests | Pending | Pending | planned |
| Durable destination and readback proof | Pending: API destination and missing-readback tests | Pending | Pending | planned |
| Compact pair artifact and final certification | Pending: pair-set resolver and aggregation test | Pending | Pending | planned |

## Commands

Focused tests will use `go test -timeout 20m ./cmd/connectorgen`; generator
checks will use `go run ./cmd/connectorgen certification-matrix --check`.
No test may contact a real provider or use real credentials.
