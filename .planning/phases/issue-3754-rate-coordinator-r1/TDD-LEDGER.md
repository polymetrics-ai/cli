# TDD ledger — issue #3754

Manual-GSD fallback: the requested phase is absent from the numerical roadmap, so generated
GSD prompts are executed inline and this ledger is the durable red/green record.

| ID | Requirement | RED evidence to retain | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | Default process-local protection is real and labelled honestly | Existing declaration has no explicit coordination mode/provenance | Local limiter shares one opaque scope in-process; visible safe status says `process-local` and never `shared` | Planned |
| R2 | `require_shared` is explicit and never inherited | Absent/invalid coordination declaration is accepted or endpoint configuration selects shared by itself | Schema/semantic table tests prove local default and only declared `require_shared` selects shared | Planned |
| R3 | Shared coordinator grants atomically from server-time TTL state | Two clients can both consume a one-unit shared scope or no shared registry type exists | Atomic grant/block, reset expiry, context cancellation, and typed unavailable reason pass | Planned |
| R4 | Require-shared fails closed | Resolver uses local limiter after shared open/ping/admission failure | `errors.As` reaches typed reason naming missing coordinator; no request is sent | Planned |
| R5 | Scope identity remains secret-free | No cross-registry test proves raw subject/key absence | Test canary scope/binding material is absent from public status/error/storage key, files, logs, receipts, and delivery evidence; type path accepts only #3863 opaque scope key | Planned |
| R6 | Two processes obey one budget when shared is engaged | No cross-process real coordinator test exists | Opt-in Dragonfly test launches two helper processes under one opaque key; exactly one grant succeeds per window | Planned |

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
