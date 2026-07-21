# Summary — Issue #475

Status: planning complete; RED tests next.

This phase owns the scoped in-process Pi AgentSession role runtime, least-authority tool policy,
trusted role prompt envelopes, and bounded redacted handoffs. Production implementation has not
started. The healthy repo-local GSD adapter does not expose `programming-loop`, so the phase is
executing the same lifecycle under an explicit `manual_gsd_fallback`.

No behavior, dependency, external state, credential, Git/GitHub state, or parent artifact has been
changed at this checkpoint.
