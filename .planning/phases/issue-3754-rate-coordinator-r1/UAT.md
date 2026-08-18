# UAT — correction 3/5, issue #4035

`verify-work 4035` was executed inline because #4035 is not a registered numerical roadmap phase and the canonical contract forbids spawning workflow roles. All deliverables are automated and have no human-judgment dependency.

| Deliverable | Automated probe | Result |
| --- | --- | --- |
| Late observation retains its budget effect after lease expiry | `TestRateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation` | Pass — a second grant proves released occupancy; a late 429/reset then produces a one-minute admission refusal. |
| Stalled UDS I/O stops on caller cancellation | `TestUnixRateBudgetCoordinatorClientCancellationInterruptsStalledExchange` | Pass — the accepted, stalled connection returns `context.Canceled` promptly. |
| A post-cancellation response is not a grant | `TestUnixRateBudgetCoordinatorClientCancellationWinsResponseRace` | Pass — the peer writes a valid ready response only after cancellation and the caller still receives `context.Canceled`. |
| Separate processes share the small budget and cleanup safely | `TestUnixRateBudgetCoordinatorMultiProcessTinyBudget` | Pass — 3 grants, 5 refusals, 0700/0600 permissions, and absent socket/run directory after owner close. |

The tests ran through `go test -race -timeout 20m ./internal/coordination/... ./internal/connectors/engine/...`; no credentials or external provider calls were used.
