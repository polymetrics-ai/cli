# Issue #4072 discussion record

**Date:** 2026-08-14

**Scope:** Gate GitHub App installation-token minting through shared rate
admission without changing the #3754 physical-route boundary.

| Decision | Result |
|---|---|
| Admission point | Keep #3754's `Requester.Do` physical-send admission. |
| Hook capability | Provide only a private, engine-owned declared-route JSON requester; do not expose a coordinator or raw transport. |
| Failure behavior | Missing or unreachable `require_shared` coordination returns the existing typed unavailable error before transport. |
| Live proof | Two isolated test processes contend for a one-request budget in real local Dragonfly; observable token POST count must be one. |
| Deferred work | Redirect admission remains #4119's accepted pre-redirect residual; no CLI, provider, policy, or PR work is in this slice. |

No user choice was needed: the issue and landed #3754/#4122 contract fix the
architecture.
