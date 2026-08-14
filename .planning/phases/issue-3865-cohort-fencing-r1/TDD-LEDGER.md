# TDD LEDGER — issue #3865 verified-auth cohort fencing

Manual-GSD fallback: #3865 is not a numbered roadmap phase. Generated GSD prompts are executed inline by the canonical single worker; this ledger is the durable RED/GREEN record.

| ID | Requirement | RED evidence | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | A closed typed outcome allows only verified invalid authentication to fence | Captured: focused test failed to compile because the outcome and coordinator contracts do not exist | Test proves timeout, transport/provider, unverified, unknown, and unrecognized outcomes leave admission available; verified invalid alone changes health | Green |
| R2 | Fencing cancels siblings and prevents all later admissions/sends | Captured: member/admission API is absent | Race test waits for cancellation, verifies a post-fence admission failure, and asserts `sends == 0` | Green |
| R3 | Cohorts isolate and repair starts a new epoch without losing audit evidence | Captured: epoch/repair APIs are absent | Test asserts other cohort sends once, repaired epoch increases, persisted `LastFencedEpoch` retains the previous epoch, stale member is rejected without sending, and fresh member sends once; an unverified repair leaves the fence in place | Green |
| R4 | Restart reloads a fence deterministically and concurrent post-fence admission is fail-closed | Captured: opaque health-store persistence/reload seam is absent | Under `-race`, a new coordinator from the same opaque state store accepts zero post-fence admissions/sends and reports the stale epoch error | Green |

## RED command log — 2026-08-15

Production coordinator code did not exist when this command ran.

```text
$ go test -count=1 -run '^TestAuthCohortCoordinator' ./internal/coordination
# polymetrics.ai/internal/coordination [polymetrics.ai/internal/coordination.test]
internal/coordination/auth_cohort_test.go:30:11: undefined: AuthenticationOutcome
internal/coordination/auth_cohort_test.go:32:30: undefined: AuthenticationOutcomeUnknown
internal/coordination/auth_cohort_test.go:33:41: undefined: AuthenticationOutcomeUnverifiedInvalid
internal/coordination/auth_cohort_test.go:34:40: undefined: AuthenticationOutcomeTransportFailure
internal/coordination/auth_cohort_test.go:35:39: undefined: AuthenticationOutcomeProviderFailure
internal/coordination/auth_cohort_test.go:36:39: undefined: AuthenticationOutcomeVerifiedHealthy
internal/coordination/auth_cohort_test.go:41:19: undefined: NewAuthCohortCoordinator
internal/coordination/auth_cohort_test.go:41:44: undefined: NewMemoryAuthCohortHealthStore
internal/coordination/auth_cohort_test.go:65:17: undefined: NewAuthCohortCoordinator
internal/coordination/auth_cohort_test.go:65:42: undefined: NewMemoryAuthCohortHealthStore
internal/coordination/auth_cohort_test.go:65:42: too many errors
FAIL	polymetrics.ai/internal/coordination [build failed]
FAIL
```

The failure proves the test is wired to the intended observable API and could not pass against the pre-feature coordinator: no typed outcome, admission, cancellation, health store, repair, or epoch seam existed.

## GREEN command log — 2026-08-15

```text
$ go test -count=1 -run '^TestAuthCohortCoordinator' ./internal/coordination
ok      polymetrics.ai/internal/coordination

$ go test -race -count=1 -timeout 20m ./internal/coordination/...
ok      polymetrics.ai/internal/coordination
ok      polymetrics.ai/internal/coordination/issueguard
```

`TestAuthCohortCoordinator_VerifiedFailureCancelsSiblingsAndRejectsNewAdmissions`
waits for an in-flight sibling's context cancellation and records zero sends after it is
fenced. `TestAuthCohortCoordinator_RestartAndRaceNeverAdmitAFencedCohort` reconstructs a
coordinator from its persisted opaque health store, attempts 64 concurrent admissions after a
verified fence, and asserts the send counter remains zero before proving repair advances the epoch
and makes the pre-restart member stale.
