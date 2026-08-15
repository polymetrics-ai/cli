# TDD Ledger: live-path defects r1

## Planned test contract

| Issue | Class | Named evidence and observable |
| --- | --- | --- |
| #4119 | Happy | Redirect to a declared `require_shared` destination is admitted/charged by that destination's policy, with exactly the expected provider sends. |
| #4119 | Bad | Canonical destination that cannot be admitted returns `*coordination.SharedRateLimitUnavailableError` naming the destination and sends zero requests. |
| #4119 | Edge | Local-only redirect, base-prefixed route, and unchanged direct route prove canonicalization neither blocks local traffic nor charges the original route. |
| #4125 | Happy | The maximum accepted window produces the exact declared duration/TTL contract. |
| #4125 | Bad | Negative and one-past-maximum windows return the specific typed validation error before cache/coordinator I/O. |
| #4125 | Edge | Zero and a duration-overflow-sized positive window receive named typed outcomes without panic or silent clamp. |
| #4169 | Happy | A provider 401 through the production CLI construction path is a typed credential/authentication outcome whose message omits credential contents. |
| #4169 | Bad | A true internal failure remains the typed `internal_error` outcome and is not collapsed into a credential error. |
| #4169 | Edge | Rejected credential leaves provider writes at zero and checkpoint state unchanged; adjacent statuses retain their defined categories. |

## Actual evidence

### 2026-08-16 — planning checkpoint

- Red: pending. No production paths have been edited.
- Green: pending.
- Manual GSD fallback: prompts resolved and executed inline because no
  compatible isolated GSD worker is available and the task forbids role
  spawning.
