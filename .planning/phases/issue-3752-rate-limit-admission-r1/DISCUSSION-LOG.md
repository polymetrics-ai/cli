# Discussion log — issue-3752-rate-limit-admission-r1

## GSD discussion execution

`scripts/gsd doctor` passed. `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
`execute-phase`, `verify-work`, and `code-review`. The generated command
`scripts/gsd prompt discuss-phase 3752` was applied inline.

The repository's canonical single-worker contract and the active task both forbid role spawning.
Pi's interactive runtime is not available in this worker, so this is the documented manual-GSD
fallback from `.agents/agentic-delivery/references/gsd-pi-adapter.md`: decisions are recorded here
instead of reopening issue-fixed choices through an interactive questionnaire.

## Decisions already fixed by the issue tree

| Area | Decision | Source |
| --- | --- | --- |
| Work order | #3751 declaration contract before #3752 requester core | task brief; #3751 |
| Provider evidence | policy source URL + retrieval date are mandatory | task brief; #3751 |
| Policy shapes | endpoint/tier/auth selectors, burst/sustained, cost budgets | #3751; #3750 |
| Existing fleet | optional migration; no 550-bundle rewrite | #3751 |
| Retry wait | preserve provider reset exactly; no 30-second cap | #3752; #3750 |
| Fallback retry | bounded full jitter only when no deterministic provider reset exists | #3752 |
| Typed result | `RateLimitError` wraps safe `HTTPError` and exposes reset | #3752 |
| Admission placement | directly before each logical `Client.Do` attempt and permitted redirect hop; safe replayable reads may replay internally, while strict non-idempotent writes cannot | #3752 |
| Scope privacy | never form a key from a secret; subject is non-secret | task brief; #3754 |
| Deferred work | resolver #3753, registry #3754, operator output #3755 | task brief |

No product ambiguity remains. The implementation will use only fixtures and `httptest`; it will not
make a provider request or read a credential.
