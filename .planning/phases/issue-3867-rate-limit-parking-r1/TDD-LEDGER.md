# TDD LEDGER — issue #3867 rate-limit parking and automatic resumption

Manual-GSD fallback: #3867 is not a numbered roadmap phase. Generated GSD
prompts execute inline under the canonical single-worker contract; this ledger
is the durable RED/GREEN record.

| ID | Requirement | RED evidence | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | A typed rate-limit error with reset evidence persists a closed `parked_rate_limit` record | Captured: focused tests fail to compile because the parking state, store, and engine bridge do not exist | Pending | Red |
| R2 | Same scope has zero pre-reset sends and unrelated scope continues | Captured: focused test fails to compile because parking admission is absent | Pending | Red |
| R3 | Restart resumes once from the exact committed checkpoint without replay | Captured: focused test fails to compile because re-arm/scheduler API is absent | Pending | Red |
| R4 | Cancellation, duplicate observations, callback failure, and races do not create extra sends/mutations | Captured: focused test fails to compile because lifecycle state machine is absent | Pending | Red |
| R5 | Park/resume events carry actual typed reason and reset timestamp | Captured: focused engine test fails to compile because no parking event bridge exists | Pending | Red |

## RED command log — 2026-08-15

Production parking code did not exist when this command ran.

```text
$ go test -count=1 -run '^(TestRateParkingCoordinator|TestParkRateLimitedRun)' ./internal/coordination ./internal/connectors/engine
# polymetrics.ai/internal/coordination [polymetrics.ai/internal/coordination.test]
internal/coordination/rate_parking_test.go:19:11: undefined: NewMemoryRateParkingStore
internal/coordination/rate_parking_test.go:24:11: undefined: NewRateParkingCoordinator
internal/coordination/rate_parking_test.go:24:37: undefined: RateParkingCoordinatorOptions
internal/coordination/rate_parking_test.go:28:33: undefined: ParkedRateLimitRun
internal/coordination/rate_parking_test.go:36:50: undefined: RateParkingRequest
internal/coordination/rate_parking_test.go:214:11: undefined: RateParkingEvent
internal/coordination/rate_parking_test.go:217:65: undefined: RateParkingEvent
internal/coordination/rate_parking_test.go:223:47: undefined: RateParkingEventType
internal/coordination/rate_parking_test.go:223:69: undefined: RateParkingEvent
internal/coordination/rate_parking_test.go:250:76: undefined: RateParkingTimer
FAIL    polymetrics.ai/internal/coordination [build failed]
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/rate_limit_parking_test.go:26:15: undefined: ParkRateLimitedRun
internal/connectors/engine/rate_limit_parking_test.go:71:23: undefined: coordination.RateParkingRequest
FAIL    polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

The failure proves the tests invoke missing observable behavior: a durable
parking state/store, same-scope admission, scheduler/restart lifecycle,
truthful event type, and engine typed-error bridge cannot pass accidentally.

## GREEN command log — pending

After the smallest implementation slice, record the focused package and
required race-command output here. Each test must assert the described state
transition or zero side effect.
