# DISCUSSION LOG — #3865/#3867 durable production wiring

- The captain supplied autonomous defaults and required no human checkpoint.
- Existing implementations and their merged GSD artifacts were read first.
- Re-audit result confirmed: only `MemoryAuthCohortHealthStore` and
  `MemoryRateParkingStore` exist; production code has no coordinator callers.
- Existing durable-state choice: `internal/state.JSONStore` provides a locked
  read/update, atomic file replacement, fsync, typed commit uncertainty, and no
  new runtime dependency. PostgreSQL is used for live auth/refusal evidence,
  not as a newly mandatory default control-plane service.
- Production composition reference: #4090's registry is built from
  definitions during `app.Open`, preflighted, and dispatched without a
  `_test.go`-only component construction. This task follows the same shape for
  both coordination owners.
- Edge coverage is binding: cancellation, crash mid-operation, empty/single/
  large state, duplicate/out-of-order updates, schema drift, filesystem and
  database permission/auth refusal, same-target races, interrupted resume, and
  acknowledged replay are all explicit TDD rows.
