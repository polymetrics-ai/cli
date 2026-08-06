# TDD ledger — issue-3753-rate-limit-enforcement-r1

## Red/green evidence

| Requirement | Red proof | Green proof | Status |
| --- | --- | --- | --- |
| Opaque local scopes | `go test ./internal/coordination -run 'TestRateLimitRegistry' -count=1` failed before implementation because `RateLimitKey`, `NewRateLimitRegistry`, and `RateLimitObservation.Cost` did not exist | `TestRateLimitScopeRegistrySharesLinkedBindingAndIgnoresCredentialRevision` proves linked binding shares while revision changes do not reset; an unlinked binding receives a separate budget | Green |
| Cancellable pacing | No policy limiter or context-aware injected clock existed | `TestRateLimitRegistryWaitHonorsContextCancellation` and `TestRateLimitRegistryCancelsWhileWaiting` pass with no wall-clock sleep | Green |
| All budgets/cost | No declaration-backed dynamic budget state existed | fixed/sliding/token/leaky tests, point-cost tightening, reset, observed limit, and concurrent burst/sustained budgets pass through the injected clock | Green |
| Selector resolution | No declared policy reached a requester | fixture tests prove exact endpoint + tier + auth matching, unsupported scope refusal, and `unknown`/`not_applicable`/absent no-op behavior | Green |
| Every requester path | Engine sends used `rt.Requester` directly and could bypass a resolved policy | admission tests cover check, paginated stream reads and fan-out ID fetches, direct/op reads, declarative form/multipart writes, op JSON/multipart writes, binary download, and whole-connector hook requester access | Green |
| Legacy precedence | Read had only its page-loop `base.rate_limit` pacing | `TestDeclaredRateLimitAddsToButDoesNotReplaceLegacyPageLimiter` proves the legacy sleeps and declared admission both remain independent | Green |

## Green command log

- `go test ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine -count=1`
- `go test -race ./internal/coordination ./internal/connectors/engine -run 'Test(RateLimit|DeclaredRateLimitFixture|UnknownRateLimitFixture|AbsentRateLimit)' -count=1`
- `go vet ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine`
- `go build ./cmd/pm`
