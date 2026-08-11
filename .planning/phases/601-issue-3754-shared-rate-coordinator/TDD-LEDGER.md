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
| `TestRateBudgetReserveBatchAllOrNothing` | Focused coordination test fails because no atomic `BudgetCoordinator.Decide` batch exists. | Same test proves a blocked later policy consumes neither an earlier policy nor its shared lease. | Green |
| `TestUnixRateBudgetCoordinatorMultiProcessTinyBudget` | Focused coordination test fails because there is no UDS owner/client/helper protocol. | Eight real barrier-released helpers produce exactly 3 grants and 5 typed blocks; process-local control produces 8 grants. | Green |
| `TestRequireSharedRefusesWithoutCoordinatorBeforeSend` | Focused engine test fails because backend selection is implicit local-only. | Missing `require_shared` coordinator unwraps a typed refusal and the test transport records 0 hits. | Green |
| `TestSharedRateBudgetScopesRemainIndependent` | Focused coordination test fails because no shared scope owner exists. | Same opaque scope shares a budget across clients while a distinct opaque scope grants independently. | Green |
| `TestSharedRateBudgetDeadlineTooShortDoesNotSend` | Focused engine test fails because no owner decision/deadline refusal exists. | A known wait beyond caller deadline unwraps typed deadline refusal with 0 transport hits. | Green |
| `TestSharedRateBudgetOwnerCrashFailsClosed` | Focused coordination test fails because no run epoch/owner lifecycle exists. | Owner loss returns a typed closed failure; an old epoch is refused by a fresh owner. | Green |

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

**GREEN result (2026-08-11):** the same focused command exited 0 after the
batch coordinator, UDS owner/client, and leased requester seam landed. The
coordination package executed the actual eight-helper local-process proof;
the engine package executed both zero-transport-hit refusal contracts. Follow
up coverage also passed:

```text
go test -count=1 -timeout 20m ./internal/coordination
go test -count=1 -timeout 20m ./internal/connectors/connsdk
go test -count=1 -timeout 20m ./internal/connectors/engine
go test -race -count=1 -timeout 20m ./internal/coordination
go vet ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors
go build ./cmd/pm
```

The coordinator tests additionally prove directory `0700` and socket `0600`
permissions, normal-close absence of both paths, idempotent Finish,
concurrency TTL release without a budget refund, late stricter observation
application, and mismatch rejection for a registered policy fingerprint.

## Green and refactor commands

```text
go test -count=1 -timeout 20m ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine
go test -race -count=1 -timeout 20m ./internal/coordination
go vet ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine
go build ./cmd/pm
```

## Correction rounds

5 / 5 used. Inline code review first removed unused legacy compatibility hooks
that created a second raw policy-ID keyed set beside the leased
fingerprint/scope path, then resolved new-code-only UDS cleanup/static lint
findings. #4025 is a separately owned planning-tool traceability sub-issue,
not a coordinator correction round.

Round 3 (#4035) separated concurrency expiry from late-observation retention
and made a canceled UDS caller interrupt its exchange. Round 4 (#4035) bounds
an unfinished lease's owner-time completion-observation lifetime at two minutes
from its grant. The injected-clock RED boundary is exact: before the horizon a
TTL-expired lease remains finishable and its typed tightening applies; at the
horizon an old owner retained and applied that observation and retained every
lost Finish record. GREEN removes the record before Finish lookup at or after
the horizon, leaves its consumption charged, and holds repeated lost-Finish
admission scans to one lease record.

```text
go test -count=1 -timeout 20m ./internal/coordination -run '^(TestRateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation|TestRateBudgetCompletionObservationHorizonNormalizesAtOwner|TestRateBudgetCompletionObservationHorizonAppliesLateObservationBeforeBoundary|TestRateBudgetCompletionObservationHorizonDropsAtBoundaryWithoutRefund|TestRateBudgetCompletionObservationHorizonBoundsAbandonedLeaseRecordsAndScans|TestUnixRateBudgetCoordinatorClientCancellationInterruptsStalledExchange|TestUnixRateBudgetCoordinatorClientCancellationWinsResponseRace|TestUnixRateBudgetCoordinatorMultiProcessTinyBudget)$'
```

The focused command passed, including both UDS cancellation regressions and the
eight-helper control with exactly 3 grants and 5 typed refusals.

Round 5 combines #4035 lifecycle validation and #4049 hook routing in one
source-fix/re-review cycle.

### Child #4035 — completion horizon cannot release active capacity

RED contract: defaults are applied before lifecycle validation; a normalized
completion-observation horizon at or below LeaseTTL is rejected, and direct
invalid owner state never deletes an active lease before `expiresAt`.

GREEN contract: `NewRateBudgetCoordinator` rejects those configurations,
`StartUnixRateBudgetCoordinator` returns no owner/client before any run
directory or listener exists, and cleanup retains active capacity while still
preserving the valid half-open completion horizon, charged consumptive
reservations, and bounded lost-Finish records.

### Child #4049 — require_shared hook sends resolve declarations

RED contract: a config-matching endpoint-sensitive policy can leave
`Runtime.Requester` bare or partially rate-limited, allowing a GitHub WriteHook
REST send to bypass `require_shared` when no coordinator exists.

GREEN contract: the generic default requester installs a path-resolution guard
with no transport or partial coordinator batch; each GitHub WriteHook REST send
uses `Runtime.RequesterFor` with its existing declared method/path. A missing
shared coordinator returns the existing typed
`shared_coordinator_unavailable` refusal with zero sends. GitHub
`rate_limits.json` and the POST `/graphql` exclusion remain unchanged.

**GREEN evidence (2026-08-11):** all three focused commands exited 0:

```text
go test -count=1 -timeout 20m ./internal/connectors/hooks/github ./internal/connectors/engine -run '^(TestRequireSharedGitHubWriteHookRefusesWithoutCoordinatorBeforeSend|TestPathAwareHookDefaultRequesterRefusesUnresolvedSendBeforeTransport|TestGitHubWriteHookResolvesEveryPhysicalRESTSend|TestRequireSharedRefusesWithoutCoordinatorBeforeSend|TestGitHubDeclaredRateLimits)$'
go test -count=1 -timeout 20m ./internal/coordination -run '^(TestRateBudgetObservationCleanupNeverDeletesActiveLease|TestRateBudgetCoordinatorRejectsHorizonNotGreaterThanTTL|TestUnixRateBudgetCoordinatorRejectsInvalidLifecycleBeforeListen|TestRateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation|TestRateBudgetCompletionObservationHorizonNormalizesAtOwner|TestRateBudgetCompletionObservationHorizonAppliesLateObservationBeforeBoundary|TestRateBudgetCompletionObservationHorizonDropsAtBoundaryWithoutRefund|TestRateBudgetCompletionObservationHorizonBoundsAbandonedLeaseRecordsAndScans|TestUnixRateBudgetCoordinatorClientCancellationInterruptsStalledExchange|TestUnixRateBudgetCoordinatorClientCancellationWinsResponseRace|TestUnixRateBudgetCoordinatorMultiProcessTinyBudget)$'
go test -race -count=1 -timeout 20m ./internal/coordination
```

The selected coordinator contract includes the real eight-helper control with
exactly 3 shared grants and 5 typed refusals plus the 8-grant/0-refusal
process-local control.
