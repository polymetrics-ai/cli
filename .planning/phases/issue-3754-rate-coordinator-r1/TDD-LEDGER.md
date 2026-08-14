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
