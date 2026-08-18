# PostgreSQL certification profile — context

## Locked decisions

- Target exactly one connector: `postgres`. No shared connector generator or unrelated connector definition may change in this lane.
- The task is a stacked delivery slice: `Refs #4015`, base `integration/4015-mvp-flat-r1`, landing path `integration/4015-mvp-flat-r1 → main`.
- Certification must exercise PostgreSQL through the shipped CLI against the shared local container runtime. A fixture, a hand-constructed component test, or a declaration-only matrix is supporting evidence, not the live proof.
- Every result must be truthful. A genuinely unavailable certification-harness capability is reported as `blocked` or `skipped` with the concrete reason; it is never promoted to `pass`.
- PostgreSQL has no REST direct-read or binary-download surface. Its profile must describe database-native catalog/read and managed-target transport evidence rather than fabricate API commands or `writes.json` actions.
- The six admitted sync modes are `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, and `incremental_dedupe_history`.
- The required failure demonstration must corrupt a generated/profile-driven assertion after the profile/schema compiles, make certification fail, then restore the exact source and show green again. The sabotage must not be committed.

## Discussion record

`scripts/gsd prompt discuss-phase 4015 --auto` was resolved and executed inline. The launch brief and issue #4015 settle scope, delivery base, safety boundaries, and acceptance criteria, so no further product decision is needed. This is an explicit non-interactive/manual fallback because the current Codex runtime must not spawn the GSD roles; the task directs autonomous execution.

## Initial finding to validate

The existing `CertificationSpec` only has `direct_read_candidates`, `binary_candidates`, and `write_pairings`. PostgreSQL's useful surface is native database catalog/read, a polling-watermark source, and managed-target apply strategies. The current API-oriented fields cannot truthfully encode that surface. The implementation must either add a definition-owned database certification shape and run it, or leave any irreducible stage blocked with the precise harness limitation.
