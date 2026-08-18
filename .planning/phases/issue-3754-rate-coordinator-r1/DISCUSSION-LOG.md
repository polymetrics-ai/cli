# Discussion log — issue #3754

Mode: generated `scripts/gsd prompt discuss-phase 3754 --auto`; executed inline because
issue #3754 is not a numbered roadmap phase and the task is explicitly autonomous.

| Area | Auto-selected decision | Why |
| --- | --- | --- |
| Default coordination | Process-local registry | Dependency-free CLI default must remain truthful and useful. |
| Shared selection | Explicit `require_shared` declaration only | An endpoint/config presence must not silently upgrade an unrelated policy. |
| Shared failure | Typed fail-closed error | Downgrading would falsely claim account-wide coordination. |
| Identity | Existing opaque rate-scope key only | #3863 is the sole secret-free identity contract. |
| External state | Atomic ephemeral Dragonfly state using server time/TTL | Runtime architecture assigns counters/leases to Dragonfly and forbids durable truth. |
| Deferred work | Parking, resume, provider-specific budgets | Owned by #3867 and #3990 respectively. |

No deferred capability is implemented by this issue.

## Correction 5/5 — #4049

| Area | Decision | Why |
| --- | --- | --- |
| Default requester | Fail closed only when a selector requires method/path resolution and its static configuration matches. | A direct caller cannot truthfully determine an endpoint or exclusion selector and must not send or consume a mixed-policy admission. |
| Declared requester | Keep policy admission at the physical send boundary. | Redirect and escaped-path controls must still be evaluated on the actual request, not only the declared template. |
| GitHub hook routing | Map every direct REST send to a pre-existing action or API-surface declaration. | This restores the engine-owned boundary without creating generic HTTP-write surface. |
| Deferred work | No rate policy, GraphQL, auth, CLI/docs, or coordinator changes. | Those are outside #4049; #4125, #4136, and #4090 remain separately tracked. |
