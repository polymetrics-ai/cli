# Human decision

A human decision is explicit, time-bounded, run/generation/gate-bound, and one-shot. Agents, answer
defaults, prior conversation, and inferred intent cannot create or consume it. Resuming a blocked
run consumes the decision, increments generation, and fences all old leases, grants, and evidence.
