# Connector release certification: prove API and PostgreSQL through DuckDB/Parquet

## Goal

Before the next PM release, prove one complete API-connector flow and one complete database-connector flow through the common Parquet warehouse mediated by embedded DuckDB. GitHub is the API connector and PostgreSQL is the database connector.

This issue integrates the existing GitHub certification parent #3988 and PostgreSQL parity parent #3972. It does not replace their issue/sub-issue ownership. Shared any-to-any transport dependencies remain owned by #3862, especially #3864 and #3866.

Target completion: 2026-08-12.

## Required executable legs

1. GitHub API source -> DuckDB/Parquet warehouse.
2. DuckDB/Parquet warehouse -> GitHub API destination.
3. PostgreSQL source -> DuckDB/Parquet warehouse.
4. DuckDB/Parquet warehouse -> managed PostgreSQL destination.

At least one cross-connector route must also prove that the warehouse is the mediator rather than connector-specific glue.

## PM acceptance surface

The integrated current-SHA binary must execute and produce evidence for:

- direct read;
- direct write;
- ETL;
- reverse ETL;
- authored flow execution;
- an actually fired schedule;
- bounded read-only database query.

## Sync contract

Certification must measure exact runtime admission and execution for every applicable canonical mode: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, `incremental_dedupe`, `incremental_dedupe_history`, and source-only `change_capture`. `full_refresh` is the source semantic used by the two full modes, not an additional canonical mode.

Unsupported or non-applicable destination semantics must produce a typed, evidence-bearing result. They must never be inferred as implemented from a broad `read` or `write` capability.

## Warehouse correctness

Each applicable passing cell must prove immutable workset identity, WAL durability, DuckDB materialization/query, Parquet row and content-hash readback, checkpoint ordering, destination receipt/readback, restart or replay behavior, cleanup, and the exact tested commit SHA.

## Delivery topology

- This issue owns one combined draft integration PR to `main`.
- #3988 owns a GitHub parent branch and draft PR into the combined integration branch.
- #3972 owns a PostgreSQL parent branch and draft PR into the combined integration branch.
- Existing child and nested issues retain one implementation PR each, targeting their corresponding parent branch.
- Existing #3993 and #4014 work must be recovered and split/reused; do not discard or duplicate it.
- Child PRs merge in dependency order into their parent, then both connector parents merge into the combined integration branch.
- Only the final current-SHA integration PR is eligible for captain-authorized merge to `main`.

## Assembly-unit tracker

This section records the current production-construction assembly unit without changing historical issue truth or claiming certification.

- [x] #4081 — `feat(sync): construct demonstrable warehouse-mediated transport` is the single direct child that owns the narrow GitHub source → durable warehouse → independent reopen → typed GitHub destination walking slice.
- [ ] #4081 is planning-only until remote `docs/4015-connector-release-certification` contains the corrected #4019 Transport parent, including #4079 at `aaf288d069adc1b67a09500afcca4be4a6d1bab3`.
- [ ] Its draft child PR must target `docs/4015-connector-release-certification`, never `main`, and must retain refs to #4081, #4015, #3862, and #4077.
- [ ] The slice must prove real-binary durable stage/reopen, durable destination receipt before checkpoint CAS, independent read-back, typed inverse cleanup, and zero residue before it is considered demonstrable.

## Validation posture

The real functional acceptance test is the primary gate. Use no-mistakes for bounded delivery, with at most five correction/validation cycles per PR. Do not spend those cycles on non-blocking polish while an executable acceptance path remains red. Preserve and escalate a concrete runtime blocker if five cycles are exhausted.

## Completion

Complete only when the combined branch demonstrates all required applicable cells and PM functions through the common DuckDB/Parquet middleware, publishes machine-readable evidence, leaves no test residue, and passes final CI/no-mistakes validation.
