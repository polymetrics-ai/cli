---
coverage:
  - id: D1
    description: Typed authoritative rate-limit evidence is persisted as a truthful parked outcome across restart.
    verification:
      - kind: unit
        ref: internal/coordination/rate_parking_test.go:TestRateParkingCoordinator_PersistsAcrossRestartAndResumesOnlyAfterReset
        status: pass
      - kind: unit
        ref: internal/connectors/engine/rate_limit_parking_test.go:TestParkRateLimitedRun_PersistsOnlyTypedAuthoritativeEvidence
        status: pass
    human_judgment: false
  - id: D2
    description: A parked scope admits zero sends before reset while unrelated scopes remain available.
    verification:
      - kind: unit
        ref: internal/coordination/rate_parking_test.go:TestRateParkingCoordinator_ConcurrentSameScopeAdmissionHasZeroSends
        status: pass
      - kind: unit
        ref: internal/coordination/rate_parking_test.go:TestRateParkingCoordinator_IsolatesScopesAndMakesDuplicateCancellationAndFailureObservable
        status: pass
    human_judgment: false
  - id: D3
    description: Automatic resumption starts at or after reset from the exact committed checkpoint and does not replay acknowledged apply work.
    verification:
      - kind: unit
        ref: internal/coordination/rate_parking_test.go:TestRateParkingCoordinator_PersistsAcrossRestartAndResumesOnlyAfterReset
        status: pass
      - kind: other
        ref: go test -race -count=1 -timeout 20m ./internal/coordination/... ./internal/connectors/engine/...
        status: pass
    human_judgment: false
  - id: D4
    description: Duplicate parking, cancellation, and callback failure have explicit durable and zero-send outcomes.
    verification:
      - kind: unit
        ref: internal/coordination/rate_parking_test.go:TestRateParkingCoordinator_IsolatesScopesAndMakesDuplicateCancellationAndFailureObservable
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3867 rate-limit parking and automatic resumption

Implemented a connector-neutral rate-limit parking coordinator and engine bridge.

- Only a typed `*connsdk.RateLimitError` with an authoritative reset instant and known source can create a `parked_rate_limit` record; generic, unknown, and reset-less errors make zero parking mutations.
- The stored record contains only the opaque rate-limit scope, a cloned committed checkpoint, reset instant, and typed reason. Reconstructing the coordinator restores the record and re-arms its timer.
- Same-scope admission is refused until a successful automatic resume at or after reset; unrelated scopes remain admissible. The restart proof verifies the exact committed checkpoint reaches the resume callback with no replayed apply.
- Duplicate observations schedule once, cancellation produces zero callbacks, and a failed resume retains parked state. Secret-free events retain the actual typed reason and reset timestamp.

The generated GSD `execute-phase`, `verify-work`, and `code-review` prompts were executed inline under the documented single-worker fallback. `TDD-LEDGER.md`, `VERIFICATION.md`, `UAT.md`, and `REVIEW.md` contain the RED/GREEN, verification, acceptance, and review records.
