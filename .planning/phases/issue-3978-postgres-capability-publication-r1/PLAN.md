# Plan — Issues 3978 and 3977: PostgreSQL capability publication

## Goal

Make the already-proven PostgreSQL CDC executor reachable through the production application and connection-owned warehouse, then publish exact `write=true`, `cdc=true`, `query=false` metadata with all generated surfaces in parity.

## TDD slices

1. **Red — application dispatch.** Add an app test whose implemented changefeed emits one committed transaction. `RunETL(change_capture)` must currently fail with `ModeNotExecutableError`; the test will require observable Parquet/WAL rows and a committed full checkpoint instead.
2. **Green — committed transaction port.** Add the smallest connector-level streaming transaction receiver/receipt contract. Adapt PostgreSQL's committed transaction stage to use it when supplied while preserving the existing connector-local callback path.
3. **Green — connection warehouse receiver.** Route `change_capture` only when the source has a matching implemented changefeed and the destination materializes a local warehouse. Atomically publish a complete transaction to the owned WAL, rebuild the single Parquet table, persist the checkpoint with the warehouse acknowledgement, then return so PostgreSQL may acknowledge the LSN. Refuse missing descriptors, non-warehouse targets, cursor fallback, and receipt/checkpoint failures before an LSN can advance.
4. **Red/green — publication.** Update PostgreSQL source metadata and native override to `write=true`, `cdc=true`, `query=false`; update exact behavior tests. Regenerate connector catalog, docs, website data, and golden transcripts with repository generators.
5. **Live proof.** Run the PostgreSQL dbtest suite through the explicit Colima socket. Assert existing CDC tests observe committed rows/checkpoints/acknowledged LSNs and existing managed-target tests observe actual row mutations and durable receipts. Run focused application tests proving the dispatch bridge.
6. **Verify and review.** Run formatting, focused/race tests, build/vet, connector/docs gates, and deep inline code review. Record base-only #4158 separately without modifying it.

## Red / Green evidence requirements

- Red must fail because production `RunETL` has no `change_capture` dispatch, not because a fixture is malformed.
- Green must assert actual warehouse record contents plus a committed `CheckpointEnvelope`; flag-string presence alone is insufficient.
- CDC acknowledgement ordering is warehouse transaction receipt -> app checkpoint persistence -> native LSN acknowledgement.
- Write proof is the existing live managed-target path's exact PostgreSQL row state and delivery receipt, rerun on this branch.

## CLI/docs/website parity

- `pm connectors inspect postgres --json` and `connectors catalog --capability` must agree with the metadata.
- Bare `pm connectors`, `pm help connectors`, and relevant `--help` surfaces remain unchanged but are smoke-checked.
- Connector docs, generated catalog, website generated data, and golden transcripts are regenerated or explicitly shown unchanged.

## Checkpoints

1. Planning plus red-test evidence.
2. Green application dispatch and connector transaction boundary.
3. Capability publication and generated parity.
4. Live/scoped verification and review fixes.
5. Push, open the explicit-base PR, verify its API-reported base, and report the full URL.
