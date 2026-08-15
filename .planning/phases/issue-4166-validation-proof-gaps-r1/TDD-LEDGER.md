# Issue 4166 TDD Ledger

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Lifecycle / delivery contract | The base branch lacks the newly referenced task-header template, so the template was read from the detached launch commit `b3df1c1c7`; no durable issue-4166 header existed. | `4166-CONTEXT.md` records the exact base, landing path, working branch, acceptance assertions, and live/fake reasons before production edits. | `scripts/gsd doctor`, all lifecycle `sources` commands, and `go run ./cmd/agentcontractgen check` passed. GSD prompts are executed inline because role spawning is forbidden. |
| Gap 1 write-action coverage | Pending: baseline sweep reports unpaired actions `blocked` and returns a passing stage even when the action contract is sabotaged. | Pending. | Must name the broken action and produce terminal certification failure; inventory presence is insufficient. |
| Gap 2 declared transport | Pending: baseline certification uses the legacy GitHub→warehouse materializer and does not execute the definition-owned transport pair. | Pending. | Must observe source, durable stage/reopen, destination plan/apply/read-back, and acknowledged checkpoint; preflight alone is insufficient. |
| Gap 3 flow round trip | Pending: existing external proof builds a child binary but its `flow_roundtrip` is capture→warehouse→query only and targets an existing lab repository; it does not create a disposable repo or run a connector-backed reverse action in the flow. | Pending. | Live credential variables are absent at plan time. A skipped test will be recorded as an open gap, not GREEN. |

## Red / Green Rule

Each gap needs an intact control and a negative control. The proof is GREEN only when the intact control has an observable execution/state transition and the deliberately broken condition fails at the claimed boundary.

