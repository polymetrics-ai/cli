# Summary — Issue #475

Status: RED captured; minimal GREEN implementation next.

This phase owns the scoped in-process Pi AgentSession role runtime, least-authority tool policy,
trusted role prompt envelopes, and bounded redacted handoffs. Fake-SDK authority and lifecycle
tests now fail at the expected missing-production-module boundary. The healthy repo-local GSD
adapter does not expose `programming-loop`, so the phase is executing the same lifecycle under an
explicit `manual_gsd_fallback`.

No production behavior, dependency, external state, credential, GitHub state, or parent artifact
has been changed at this checkpoint.
