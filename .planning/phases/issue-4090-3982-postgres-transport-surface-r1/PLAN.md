# Plan — PostgreSQL transport surface truthfulness

## Goal

Make PostgreSQL's advertised source/destination mode intersection exactly match the production registry's preflight admission, while keeping the generated certification shard truthful.

## TDD sequence

1. **Red — production happy path.** From a real `app.Open` registry, add a separately named table test that preflights every PostgreSQL destination mode and asserts the actual resolved source/destination executor references.
2. **Red — rejected path.** Assert the currently source-advertised but destination-unmatched `incremental_dedupe_history` mode is refused with the source-mode reason before a source executor can run.
3. **Red — declaration edge.** Assert source and destination mode sets are the same exact five-mode intersection, with `full_overwrite` retained and no duplicate/extra source-only mode.
4. **Green.** Remove only the unmatched source mode. Regenerate the PostgreSQL certification shard with its scoped generator, then prove it is byte-current with `--check`.
5. **Verify/review.** Run the focused App, connector, CLI, and generator tests; build and inspect `pm`; run the applicable individual repository gates; conduct inline verify-work and review.

## Safety fences

- No capability flip, destination `write:true`, generic SQL, credentials, provider I/O, or database integration run.
- Do not remove the currently reachable `full_overwrite`, `incremental_append`, or `incremental_dedupe` modes merely to reproduce the stale audit premise.
- Generated JSON is generator-owned; no manual artifact edit.
