# Discussion log — seven connector extraction r1

`scripts/gsd prompt discuss-phase cli-sweep-seven-connector-extract-r1 --auto` was generated and
executed inline as the non-interactive discussion record. The captain order fixed every material
choice:

- source: `c28bc75a3`, never the sweep branch worktree;
- include exactly seven named connector bundle deltas;
- exclude `github` and `zendesk-support` completely;
- regenerate, never hand-merge, operation ledger, command surface, docs, and website data;
- use no credentials and make no live calls;
- verify all implemented command paths through a built binary;
- report the not-certified / never-live-tested status in the PR handoff.

No remaining product decision is required before planning.
