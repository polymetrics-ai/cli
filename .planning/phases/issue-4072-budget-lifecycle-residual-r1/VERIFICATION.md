---
phase: issue-4072-budget-lifecycle-residual-r1
status: verified
---

# #4072 residual verification checklist

| Must-have | Result | Evidence |
| --- | --- | --- |
| Grant calls `Decide` and `Finish` once each | pass | `TestGitHubAppAuthBudgetLifecycleGrantFinishesExactlyOnce` asserts `Decide=1`, `Finish=1`, `send=1`, the granted opaque lease, and an attempted completion observation. |
| Refusal has one `Decide`, zero `Finish`, and zero sends | pass | `TestGitHubAppAuthBudgetLifecycleRefusalDoesNotFinishOrSend` asserts `Decide=1`, `Finish=0`, `send=0`, and `*RateBudgetRefusalError{Code: "reservation_denied"}`. |
| Required shared failure remains typed and pre-I/O | pass | `TestGitHubAppAuthRateAdmissionRequireSharedRefusesBeforeTokenSend` and `...Unreachable...` remained green in the non-race GitHub auth suite. |
| Failed mint still makes exactly one POST | pass | `TestGitHubAppAuthRateAdmissionDoesNotRetryTokenMint` remained green in that suite. |
| No lifecycle secret retention | pass | The fake stores only count values, `ReservationBatch`, opaque lease, and `CompletionObservation`; its type has no request/header/body/JWT/private-key/token field. The production batch is constructed only from declared policy fingerprint, opaque scope, and declared budgets. |
| Required package, consumer, generator, lint, docs, and boundary gates pass | pass | Commands and results below. |

## Commands run

| Command | Result |
| --- | --- |
| `scripts/gsd doctor` | pass |
| `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, `code-review` | pass — all resolve through the pinned adapter |
| `go run ./cmd/agentcontractgen check` | pass |
| `go test -timeout 20m ./internal/connectors/hooks/github -run '^TestGitHubAppAuthBudgetLifecycle' -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/hooks/github -run '^TestGitHubAppAuth' -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/engine -count=1` | pass |
| `go test -timeout 20m ./internal/connectors -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/connsdk -count=1` | pass |
| `go test -timeout 20m ./internal/cli -count=1` | pass |
| `go test -timeout 20m ./cmd/connectorgen -count=1` | pass |
| `go test -timeout 20m -race ./internal/connectors/hooks/github -run '^TestGitHubAppAuthBudgetLifecycle' -count=1` | pass |
| `go vet ./internal/connectors ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/hooks/github` | pass |
| `go build ./cmd/pm` | pass |
| `pnpm --dir website run gen:docs` (twice) | pass — byte-stable, no generated diff |
| `make tidy-check lint docs-check smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` | pass (each gate ran independently) |

## Disclosed non-gate result

`go test -timeout 20m -race ./internal/connectors/hooks/github -run '^TestGitHubAppAuth' -count=1` did not complete: the pre-existing unreachable-coordinator case consumed its fixed five-second context under race instrumentation and returned `context deadline exceeded` rather than its expected typed unavailable error. The focused lifecycle race test passed. The normal GitHub auth regression suite is green; no test was weakened, skipped, or changed to conceal this timing issue.
