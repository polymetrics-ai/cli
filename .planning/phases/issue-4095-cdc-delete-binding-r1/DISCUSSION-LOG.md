# Discussion log — Issue 4095

`scripts/gsd prompt discuss-phase issue-4095-cdc-delete-binding-r1 --auto`
was resolved and executed inline. The launch brief and issue lock the only
meaningful choices: PostgreSQL is the sole source, existing keyed history apply
is the destination behavior, and physical absence cannot be promoted to a
delete. No product decision remains open.

The active repository base was refreshed to `f49f9a9ee`, which contains the
native apply boundary from #4153 without the prohibited #4154 refactor.
