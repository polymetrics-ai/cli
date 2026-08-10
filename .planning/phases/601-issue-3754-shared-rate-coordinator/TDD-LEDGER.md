# Phase 601 TDD ledger — #3754 optional shared rate-budget coordinator

**Execution mode:** inline/manual GSD fallback; the canonical delivery
contract forbids GSD role spawning in this lane.

## Required skills

Loaded before planning: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, and
`golang-concurrency`. `golang-lint` remains required before code review.

## Mandatory red/green matrix

| Contract | RED command / expected missing behavior | GREEN command / proof | Status |
| --- | --- | --- | --- |
| `TestRateBudgetReserveBatchAllOrNothing` | Focused coordination test fails because no atomic `BudgetCoordinator.Decide` batch exists. | Same test proves a blocked later policy consumes neither an earlier policy nor its shared lease. | Red — missing API |
| `TestUnixRateBudgetCoordinatorMultiProcessTinyBudget` | Focused coordination test fails because there is no UDS owner/client/helper protocol. | Eight real barrier-released helpers produce exactly 3 grants and 5 typed blocks; process-local control produces 8 grants. | Red — missing API |
| `TestRequireSharedRefusesWithoutCoordinatorBeforeSend` | Focused engine test fails because backend selection is implicit local-only. | Missing `require_shared` coordinator unwraps a typed refusal and the test transport records 0 hits. | Red — missing API |
| `TestSharedRateBudgetScopesRemainIndependent` | Focused coordination test fails because no shared scope owner exists. | Same opaque scope shares a budget across clients while a distinct opaque scope grants independently. | Red — missing API |
| `TestSharedRateBudgetDeadlineTooShortDoesNotSend` | Focused engine test fails because no owner decision/deadline refusal exists. | A known wait beyond caller deadline unwraps typed deadline refusal with 0 transport hits. | Red — missing API |
| `TestSharedRateBudgetOwnerCrashFailsClosed` | Focused coordination test fails because no run epoch/owner lifecycle exists. | Owner loss returns a typed closed failure; an old epoch is refused by a fresh owner. | Red — missing API |

## Initial RED command

```text
go test -count=1 -timeout 20m ./internal/coordination ./internal/connectors/engine -run '^(TestRateBudgetReserveBatchAllOrNothing|TestUnixRateBudgetCoordinatorMultiProcessTinyBudget|TestRequireSharedRefusesWithoutCoordinatorBeforeSend|TestSharedRateBudgetScopesRemainIndependent|TestSharedRateBudgetDeadlineTooShortDoesNotSend|TestSharedRateBudgetOwnerCrashFailsClosed)$'
```

**RED result (2026-08-11):** exited 1 as expected. The coordination package
failed to compile because the contract does not yet exist:
`connsdk.ReservationPolicy`, `RateBudgetPolicyFingerprint`,
`connsdk.ReservationKey`, `connsdk.ReservationBatch`,
`NewRateBudgetCoordinator`, `RateBudgetCoordinatorOptions`, and
`connsdk.CompletionObservation` were undefined. The engine package likewise
reported the missing explicit backend/configuration and UDS contract:
`RuntimeConfig.RateBudgetBackend`,
`connsdk.RateBudgetBackendRequireShared`,
`connsdk.RateBudgetRefusalError`,
`coordination.StartUnixRateBudgetCoordinator`, and
`coordination.UnixRateBudgetCoordinatorOptions` were undefined. No test ran
because both packages stopped at the intended missing-production-API compile
boundary.

## Green and refactor commands

```text
go test -count=1 -timeout 20m ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine
go test -race -count=1 -timeout 20m ./internal/coordination
go vet ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine
go build ./cmd/pm
```

## Correction rounds

0 / 5 used. A round is counted only when a GSD verifier, equivalent Shepherd
evidence, code review, CI, or no-mistakes finding causes production/test code
to change. A newly discovered gate defect requires a #3754 sub-issue before a
code change.
