# TDD ledger — issue #3754

Manual-GSD fallback: the requested phase is absent from the numerical roadmap, so generated
GSD prompts are executed inline and this ledger is the durable red/green record.

| ID | Requirement | RED evidence to retain | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | Default process-local protection is real and labelled honestly | Existing declaration has no explicit coordination mode/provenance | Local limiter shares one opaque scope in-process; visible safe status says `process-local` and never `shared` | Green |
| R2 | `require_shared` is explicit and never inherited | Absent/invalid coordination declaration is accepted or endpoint configuration selects shared by itself | Schema/semantic table tests prove local default and only declared `require_shared` selects shared | Green |
| R3 | Shared coordinator grants atomically from server-time TTL state | Two clients can both consume a one-unit shared scope or no shared registry type exists | Atomic grant/block, server-time TTL, all four models, and context cancellation pass against a live Dragonfly service | Green |
| R4 | Require-shared fails closed | Resolver uses local limiter after shared open/ping/admission failure | `errors.As` reaches typed `coordinator_not_configured`; resolver construction fails before requester send | Green |
| R5 | Scope identity remains secret-free | No cross-registry test proves raw subject/key absence | Actual #3863 opaque key derivation is asserted against subject/binding canaries; public output/errors contain none; no argv/env/file/log/receipt path accepts raw credentials | Green |
| R6 | Two processes obey one budget when shared is engaged | No cross-process real coordinator test exists | Opt-in Dragonfly test launches two helper processes under one opaque key; exactly one grant succeeds per window | Green |
| R7 | Website generated data matches the checked-in CLI reference | CI runs `31809146629` and `31809146767` regenerate `website/lib/docs.generated.ts` after the rate-limit provenance addition | Regenerate the one checked-in artifact and re-run generator/lint/typecheck without source or dependency changes | Green |

## Red command log

No production code changed before this run.

```text
$ go test ./internal/coordination ./internal/connectors/engine ./internal/cli -run 'Test(SharedRateLimitRegistry|RateLimitRegistryStatus|RequireSharedRateLimitPolicy|LocalRateLimitPolicy|ConnectorsInspectLabelsProcessLocal)' -count=1
# polymetrics.ai/internal/coordination [polymetrics.ai/internal/coordination.test]
internal/coordination/shared_rate_limits_test.go:16:21: registry.Status undefined
internal/coordination/shared_rate_limits_test.go:26:12: undefined: NewSharedRateLimitRegistry
internal/coordination/shared_rate_limits_test.go:29:19: undefined: SharedRateLimitUnavailableError
FAIL    polymetrics.ai/internal/coordination [build failed]
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
... RateLimitPolicy has no field or method Coordination
... undefined: replaceSharedRateLimitRegistryForTest
FAIL    polymetrics.ai/internal/connectors/engine [build failed]
--- FAIL: TestConnectorsInspectLabelsProcessLocalRateLimitProtection
    inspect output did not label process-local rate-limit protection
FAIL    polymetrics.ai/internal/cli
```

This is the intended red baseline: no policy can require shared coordination, no shared registry or
typed fail-closed result exists, and the binary makes no process-local provenance statement.

## Green command log

```text
$ ./pm connectors inspect github | rg -A 2 '^RATE LIMIT COORDINATION$'
RATE LIMIT COORDINATION
  Process-local rate-limit protection coordinates this pm process only; it is not shared across processes.

$ go test ./internal/coordination ./internal/connectors/engine -run 'Test(RateLimitRegistryStatusIsExplicitlyProcessLocal|SharedRateLimitRegistryRefusesWhenCoordinatorIsMissing|RequireSharedRateLimitPolicyRefusesWithoutCoordinator)' -count=1 -v
=== RUN   TestRateLimitRegistryStatusIsExplicitlyProcessLocal
    shared_rate_limits_test.go:23: rate-limit coordination: process-local rate-limit protection; not shared across processes
--- PASS: TestRateLimitRegistryStatusIsExplicitlyProcessLocal (0.00s)
=== RUN   TestSharedRateLimitRegistryRefusesWhenCoordinatorIsMissing
    shared_rate_limits_test.go:40: require_shared result=refused reason=coordinator_not_configured
--- PASS: TestSharedRateLimitRegistryRefusesWhenCoordinatorIsMissing (0.00s)
PASS
ok      polymetrics.ai/internal/coordination
=== RUN   TestRequireSharedRateLimitPolicyRefusesWithoutCoordinator
--- PASS: TestRequireSharedRateLimitPolicyRefusesWithoutCoordinator (0.00s)
PASS
ok      polymetrics.ai/internal/connectors/engine
```

The live Dragonfly command and its full verbatim output were posted to issue #3754. It proved two
processes against one opaque budget (one grant, one context-cancelled block), every declared budget
model, and a `require_shared` requester admission. No credentialed provider call was run because
this connector-neutral foundation intentionally leaves provider policy declarations to #3990.

## Correction 3/5 — #4035 (planned)

| ID | RED condition | GREEN contract | Evidence status |
| --- | --- | --- | --- |
| C3-R1 | An expired lease is retired after its short retained window, so `Finish` silently drops a valid later 429/reset observation. | Expiry releases only concurrency occupancy. The lease record remains finishable, and a later valid observation tightens the next admission. | **RED captured** |
| C3-R2 | UDS `exchange` reads with only a static deadline, so cancellation cannot interrupt a stalled peer and a response that arrives in the cancellation race can become a ready/grant result. | A context cancellation advances the connection deadline and the client checks `ctx.Err()` after the response read; no late response reaches `Ready`/`Decide`. | **RED captured** |

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-context`, `golang-concurrency`, and `golang-testing`.

### RED command — 2026-08-14

```text
$ go test -count=1 -run 'Test(RateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation|UnixRateBudgetCoordinatorClientCancellation)' ./internal/coordination
# polymetrics.ai/internal/coordination [polymetrics.ai/internal/coordination.test]
internal/coordination/rate_budget_coordinator_test.go:12:85: undefined: connsdk.ReservationPolicy
internal/coordination/rate_budget_coordinator_test.go:15:22: undefined: RateBudgetPolicyFingerprint
internal/coordination/rate_budget_coordinator_test.go:27:17: undefined: NewRateBudgetCoordinator
internal/coordination/unix_rate_budget_coordinator_test.go:61:13: undefined: UnixRateBudgetCoordinatorClient
FAIL	polymetrics.ai/internal/coordination [build failed]
FAIL
```

This is a behavioral RED baseline: neither the lease lifecycle that can retain a completion observation nor the UDS client capable of receiving cancellation exists on the declared integration base.

### GREEN command log — 2026-08-14

```text
$ go test -count=1 -run 'Test(RateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation|UnixRateBudgetCoordinatorClientCancellation)' ./internal/coordination
ok      polymetrics.ai/internal/coordination

$ go test -race -count=1 -timeout 20m ./internal/coordination/... ./internal/connectors/engine/...
ok      polymetrics.ai/internal/coordination
ok      polymetrics.ai/internal/coordination/issueguard

$ go test -race -count=1 -timeout 20m ./internal/connectors/engine/... -run 'Test(RequireSharedRateLimitPolicy|EndpointRequireShared|RateLimit)'
ok      polymetrics.ai/internal/connectors/engine

$ go vet ./internal/coordination/... ./internal/connectors/connsdk ./internal/connectors/engine/...
$ make lint
0 issues.
```

The coordinator test advances the injected clock by three short lease TTLs, proves a second admission is granted (occupancy released), then applies the first lease's late 429/reset `Finish` and observes a one-minute refusal. The UDS tests assert prompt `context.Canceled` during a stalled exchange and when a valid ready response is sent after cancellation. The race command also runs `TestUnixRateBudgetCoordinatorMultiProcessTinyBudget`, whose eight subprocesses assert exactly three grants, five refusals, private socket permissions, and owner cleanup.
