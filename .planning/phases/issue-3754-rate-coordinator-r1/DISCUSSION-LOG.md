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
