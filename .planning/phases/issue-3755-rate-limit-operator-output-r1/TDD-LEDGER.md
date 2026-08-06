# TDD ledger — issue-3755-rate-limit-operator-output-r1

## Red/green evidence

| Requirement | Red proof to add | Green proof to add | Status |
| --- | --- | --- | --- |
| Declared selection | A declared test-only bundle's selected policy does not appear in an execution summary | one coalesced row reports policy ID, subject kind, and structural selector reason | Green: `TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle`, `TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets` |
| Honest absence | A bundle without `rate_limits.json` serializes no state or implies unlimited | summary is exactly `undeclared`, has no selected policies, and does not change requester behavior | Green: `TestRateLimitReportCallsAbsentDeclarationUndeclared`, `TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree` |
| Local pacing | Existing deterministic limiter waits are not surfaced | aggregate local pacing equals the injected clock's actual wait without per-request output | Green: `TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets` |
| Provider facts | Typed headers/429 are consumed by enforcement but not returned to the operator | latest remaining budget/reset-safe facts and 429 observed/honoured wait are included in the bounded summary | Green: `TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets` |
| Ordinary latency | Operator cannot distinguish a normal slow requester from pacing/provider waits | requester latency is aggregated separately from local pacing and provider retry wait | Green: `TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle` |
| Human and JSON ETL run | `pm etl run` has only record counts and JSON omits report data | human output labels declaration/pacing/provider/latency; JSON exposes identical `run.rate_limit` data | Green: `TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle`, `TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree` |
| Secret-free invariant | A report could capture a credential, raw binding, runtime subject, opaque scope, or `CredentialRevision` | regression tests seed distinct sentinel values into every prohibited field and assert neither `json.Marshal(report)` nor human output contains any | Green: `TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets`, `TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree` |
| Bounded output | A long run could append an event per request | repeated policy/admission/observation calls leave one policy row and scalar aggregates only | Green: `TestRateLimitReportCoalescesLongRunsIntoBoundedPolicySummary` |

## Expected command log

- `go test ./internal/connectors ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine ./internal/app ./internal/cli -run 'Test.*RateLimit|TestETL' -count=1`
- `go test -race ./internal/connectors ./internal/connectors/engine ./internal/app -run 'Test.*RateLimit' -count=1`
- `go vet ./internal/connectors ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine ./internal/app ./internal/cli`
- `go build ./cmd/pm`
