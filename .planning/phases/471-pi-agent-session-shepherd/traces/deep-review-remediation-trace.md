# Deep-review remediation trace

Date: 2026-07-21
Reviewed head: `7f745427d38995940b8f57517d0241d1e10d3f64`
Review artifact: `../471-REVIEW.md`
Integrated remediation head: `dcc3829d`

The mandatory GSD deep review reported ten critical defects and six warnings. None was waived.
Three non-overlapping `gpt-5.6-sol` high implementation lanes corrected the SDK lifecycle,
state/lease boundary, and controller/extension boundary with test-first evidence. Root integration
then corrected one strict-TypeScript descriptor-buffer mismatch and added a missing regression for
two launches racing while canonical worktree resolution was pending.

Final focused evidence at the integrated checkpoint:

- `node --test .pi/extensions/shepherd/*.test.ts`: 82/82 pass
- strict TypeScript no-emit over all production Shepherd modules: pass
- offline Pi 0.80.6 RPC command discovery: pass
- `git diff --check`: pass

The release candidate now rejects non-`stop` Pi terminal outcomes, preserves persisted PR identity
on resume, canonicalizes worktree aliases, uses structured sibling cancellation and joining,
linearizes terminal ownership, owns late SDK creation through cleanup, atomically appends fenced
lease transitions, rejects symlinked or escaped state, persists no arbitrary provider/model text,
validates cross-field state invariants, shares close settlement, propagates shutdown failures, and
handles native Windows absolute path forms. A fresh PR #438 canary and exact-head review remain
mandatory because the earlier canary predates this remediation.
