# Foundation issue — database certification execution and truthful roll-up

## Problem

`allStagesPassed` treats every `skipped:` stage as benign. PostgreSQL's
`declared_transport_pair` currently reports `skipped: connector has no
executable certification adapter for its declared transport pair`, so the
report can pass although its declared source/destination transport was never
executed.

## Scope

- Classify benign environmental skips separately from unexecutable applicable
  stages and make the latter fail the certification roll-up.
- Add a connector-neutral database transport certification adapter that runs
  PostgreSQL's declared polling-watermark source and managed-target
  destination against isolated live databases.
- Record accepted, redacted database proof so the generated PostgreSQL matrix
  marks only actually executed live cells.

## Acceptance

- A `--full` PostgreSQL certificate executes its declared pair; an absent
  executable adapter is non-pass and returns certification failure.
- Catalog/schema discovery, typed read, polling watermark, managed target,
  every declared apply strategy, and all declared modes including
  `incremental_dedupe_history` have explicit results.
- A live database result marks only the matching generated matrix cells;
  regeneration is byte-stable.
- The allowlisted GitHub certificate retains only genuinely benign skips.

## Relationship

This is the shared foundation for #4015's PostgreSQL profile. The delivery is
one PR against `integration/4015-mvp-flat-r1`; it must distinguish generic
runner/evidence changes from PostgreSQL-owned definitions and live tests.
