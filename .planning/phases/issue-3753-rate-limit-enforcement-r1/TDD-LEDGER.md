# TDD ledger — issue-3753-rate-limit-enforcement-r1

## Planned red/green evidence

| Requirement | Red proof | Green proof | Status |
| --- | --- | --- | --- |
| Opaque local scopes | No registry exists and same policy bindings cannot share state | linked bindings share one opaque budget; copied/unlinked bindings do not; revisions do not reset | Planned |
| Cancellable pacing | No policy limiter waits through `ctx` | injected clock wait returns `context.Canceled` without a send | Planned |
| All budgets/cost | No dynamic budget state exists | fixed/sliding/token/leaky, burst+sustained, requests+points, and higher actual cost are enforced | Planned |
| Selector resolution | No declared policy reaches a requester | endpoint/tier/auth selector tests show only applicable policies attach | Planned |
| Every requester path | Engine paths make calls with no resolved admission | counter tests cover check, page read, direct read/op read, declarative + form/multipart + op writes, and binary download | Planned |
| Legacy precedence | Read has only page-loop `base.rate_limit` pacing | matching declaration is requester admission in addition to unchanged legacy wait; absent declaration is unchanged | Planned |

## Red evidence log

Pending: write and execute the focused tests before production implementation. No production files have changed at planning time.
