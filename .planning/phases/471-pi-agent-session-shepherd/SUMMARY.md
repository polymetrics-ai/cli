# Summary

Phase: `471-pi-agent-session-shepherd`
Status: in progress

Issue #471 and an isolated main-based worktree exist. The standalone architecture, SDK boundary,
manual-GSD fallback, TDD sequence, verification plan, and PR #438 read-only canary contract are
recorded. Six focused test entries plus offline Pi RPC discovery have expected pre-production RED
evidence. The deterministic core, strict command surface, allowlisted atomic state, zero-tool
AgentSession runner, exact target evidence, controller, and extension wiring are implemented.
Independent review findings became regression tests for crash resume, stop races, post-run target
changes, cross-process leasing, early cancellation, bounded shutdown, output bounds, and persisted
DTO safety. A later exact-head deep review found ten additional critical defects and six warnings;
all are now implemented test-first at checkpoint `dcc3829d`. The hardened design verifies Pi
terminal events, binds resume to the persisted PR, canonicalizes Git worktrees, joins failed lane
groups before lease release, linearizes stop, owns setup through cleanup under one deadline,
publishes a CAS-style no-follow lease journal, persists fixed summary categories, validates state
machines, aggregates shutdown failures, and supports native Windows path forms. A root integration
race test also reserves the process-wide launch slot before asynchronous worktree resolution.
The focused suite is 82/82 green, strict TypeScript passes, and Pi 0.80.6 discovers
`/pm-shepherd` offline. The earlier PR #438 canary passed at `ccf0daf3` but is now historical;
the final hardened head still requires a fresh read-only canary. Final root gates, a repeat
exact-head review, automated PR review, and the human-gated main merge remain.
