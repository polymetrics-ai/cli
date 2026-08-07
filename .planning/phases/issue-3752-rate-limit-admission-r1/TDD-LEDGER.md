# TDD ledger — issue-3752-rate-limit-admission-r1

Manual-GSD fallback: the phase used `scripts/gsd prompt discuss-phase 3752` and
`scripts/gsd prompt plan-phase 3752 --tdd` inline because this worker cannot invoke Pi and the
canonical delivery contract forbids spawning GSD roles. This ledger is the durable red/green record.

## Planned red/green evidence

| Slice | Red test / proof | Green assertion | Status |
| --- | --- | --- | --- |
| #3751 optional loader | `Load` has no `rate_limits.json` typed result; a fixture declaration cannot be loaded and validated | optional file returns nil when absent and typed data when present | Green |
| #3751 citation | table-driven declaration lacking `source.url` or `source.retrieved_at` is not rejected today | load error identifies `rate_limits.json` and the deficient citation | Green |
| #3751 selector/state | malformed endpoint/tier/auth selector, duplicate policy ID, invalid scope subject, and inconsistent unknown/not_applicable state | each is rejected before runtime; a real-shape declared policy survives typed load | Green |
| #3752 provider reset | `Retry-After: 90` + `MaxBackoff: 30s` records `30s` through injected `Sleep` | the same fixture records exactly `90s` | Green |
| #3752 reset typing | terminal 429 exposes only `*HTTPError` today | `errors.As(err, *RateLimitError)` exposes reset and still reaches `*HTTPError` | Green |
| #3752 fallback jitter | fallback retry has no jitter hook and retries in lockstep | injected full jitter is within cap; valid provider reset calls no jitter hook | Green |
| #3752 admission | no pre-send admission exists in `doWithBody` / `DoStream` | cancellation/error prevents an initial logical `httptest` request; logical retries and redirects are re-admitted, while shallow clones preserve the admission | Green |
| #3752 observation | 429 has no typed callback | observer sees status, attempted request, source, retry duration/reset; output contains no fixture secret | Green |
| #3752 retry safety | retry cap / DisableRetries are existing contracts | typed rate limiting retains retry count and no-replay behavior | Green |

## Red evidence log

No requester production code changed before this run.

```text
$ go test ./internal/connectors/connsdk -run TestRequesterHonorsProviderRetryAfterBeyondFallbackCap -count=1
--- FAIL: TestRequesterHonorsProviderRetryAfterBeyondFallbackCap (0.00s)
    http_test.go:103: provider Retry-After wait = 30s, want exact 1m30s
FAIL
FAIL    polymetrics.ai/internal/connectors/connsdk    0.278s
FAIL
```

This is the confirmed live defect: `Requester.backoff` clamps a provider's deterministic 90-second
reset to its 30-second fallback cap. The test and this evidence stay in the branch after the green
implementation lands.

Before the #3751 loader was added, its new fixture tests failed to compile because `Bundle` had no
typed `RateLimits` result:

```text
$ go test ./internal/connectors/engine -run 'TestBundleLoad(ParsesProviderCitedRateLimits|RejectsUncitedOrMalformedRateLimits)' -count=1
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/bundle_test.go:294:7: b.RateLimits undefined (type Bundle has no field or method RateLimits)
FAIL    polymetrics.ai/internal/connectors/engine [build failed]
```

After the loader/schema implementation:

```text
$ go test ./internal/connectors/engine -run 'TestBundleLoad(ParsesProviderCitedRateLimits|RejectsUncitedOrMalformedRateLimits)' -count=1
ok      polymetrics.ai/internal/connectors/engine    0.497s
```

The requester-contract tests were then added before any requester production change. They fail to
compile against the old public shape, which has no admission/observation/error/jitter seam:

```text
$ go test ./internal/connectors/connsdk -run 'TestRequester(HonorsProviderRetryAfterBeyondFallbackCap|FallbackRetryUsesBoundedFullJitter|AdmissionPreventsInitialLogicalSend|ReturnsTypedRateLimitErrorAndObservation)' -count=1
# polymetrics.ai/internal/connectors/connsdk [polymetrics.ai/internal/connectors/connsdk.test]
internal/connectors/connsdk/http_test.go:96:3: unknown field Jitter in struct literal of type Requester
internal/connectors/connsdk/http_test.go:119:51: undefined: RateLimitRequest
internal/connectors/connsdk/http_test.go:125:50: undefined: RateLimitObservation
internal/connectors/connsdk/http_test.go:181:3: unknown field Admission in struct literal of type Requester
FAIL    polymetrics.ai/internal/connectors/connsdk [build failed]
```

The green requester suite covers exact 90-second provider waits, no jitter for a valid provider
reset, capped injected full jitter only for fallback retries, admission before initial logical
JSON/form/multipart/stream sends, caller cancellation while an admission is waiting, each logical
retry or redirect, terminal typed-429 wrapping, standard limit/remaining/reset headers, and
HTTP-date preservation. Safe replayable reads retain one admission through an internal `net/http`
replay, while strict non-idempotent writes suppress replay:

```text
$ go test ./internal/connectors/connsdk -count=1
ok      polymetrics.ai/internal/connectors/connsdk    0.873s

$ go test -race ./internal/connectors/connsdk -run '^(TestRequesterHonorsProviderRetryAfterBeyondFallbackCap|TestRequesterFallbackRetryUsesBoundedFullJitter|TestRequesterAdmissionPreventsInitialLogicalSend|TestRequesterAdmissionHonorsCallerCancellationBeforeLogicalSend|TestRequesterAdmitsEachLogicalRetryAttempt|TestRequesterReturnsTypedRateLimitErrorAndObservation|TestParseRetryAfterAtPreservesProviderDate|TestRateLimitObservationParsesStandardBudgetHeaders|TestDoStreamReturnsTypedRateLimitError)$' -count=1
ok      polymetrics.ai/internal/connectors/connsdk    1.286s
```

The manual review added two declaration-validation regression cases before the validation fix:

```text
$ go test ./internal/connectors/engine -run TestBundleLoadRejectsUncitedOrMalformedRateLimits -count=1
--- FAIL: TestBundleLoadRejectsUncitedOrMalformedRateLimits
    --- FAIL: endpoint_path_cannot_carry_outer_whitespace
    --- FAIL: cost_header_must_be_an_HTTP_field_name
FAIL
```

After rejecting whitespace-deformed endpoint selectors and non-token cost header names, the focused
loader suite and the full engine package both passed. The review also added the production-embed
guard for a future real `rate_limits.json` declaration.
