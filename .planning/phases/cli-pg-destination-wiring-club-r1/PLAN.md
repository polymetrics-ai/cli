# Plan — PostgreSQL production transport wiring club

## Goal

Make the already-implemented PostgreSQL managed driver, immutable workset delivery, and gap-free bootstrap reachable from the real `pm` entry point through the definition-owned transport registry and the connection-owned warehouse, with authenticated approval and observable live edge evidence.

## TDD slices

1. **Red — destination declaration and production registration.** Add tests that open a real App registry and require PostgreSQL destination preflight plus a non-test destination constructor. Assert unknown/mismatched factory evidence, undeclared modes (especially target `change_capture` and history), and missing approval refuse before PostgreSQL or warehouse side effects.
2. **Green — exact PostgreSQL destination adapter.** Add the destination half of `sync_transport.json`, connector-local factory, and adapter. Open the native driver per call from the validated runtime; derive structural owner/target/schema/mapping, provision only after authenticated approval, consume the reopened stage Parquet through `ChangeDeliveryWorkset` for `incremental_upsert`, persist the target receipt and baseline, then independently read back before checkpoint admission.
3. **Red/green — real relation source and bootstrap bridge.** Replace the registration-only literal stream assumption with a bounded dynamic relation contract. Route full modes to the existing snapshot reader; route a first incremental-upsert run to `BootstrapCDC` and a resumed run to `ReadCDC`. Convert inserts/updates to records and deletes to key-bound tombstones transactionally, emitting a source page only from the durable checkpoint callback.
4. **Red/green — production approval and CLI surface.** Add a closed PostgreSQL transport plan/preview command and make `pm etl run` accept its stdin-only single-use approval. Bind connection/stream IDs, source and destination endpoint revisions, mode, mapping/schema, and exact destination executor. Refuse stale, mismatched, expired, concurrent, and replayed approval before provisioning or writes. Update help, docs, website, manual/skills, and golden transcripts.
5. **Red/green — live production composition.** Through App and a built `pm`, run real PostgreSQL source → connection warehouse → real PostgreSQL target. Assert rows, types, private owner/control/ledger state, warehouse WAL/Parquet/manifests, baselines, receipts, checkpoints, LSNs, and slot state. Include API-source → warehouse → PostgreSQL target if a credential-free real API route is available; otherwise identify its credential/live-system constraint explicitly in the PR table while keeping executable code coverage.
6. **Edge and recovery matrix.** Exercise cancellation, source/target process death, empty/single/large pages, duplicates/out-of-order transactions, schema drift, auth/permission refusal, concurrent same-target runs, restart/resume, and acknowledged replay. Every refusal asserts a typed error plus unchanged target row count, control/ledger count, baseline identity, and source checkpoint.
7. **Verify/review/delivery.** Run focused tests, race tests, the mandated databaseintegration command, build the real binary, regenerate every derived artifact in one pass, run CI drift/gate commands, execute inline `verify-work` and `code-review`, inspect capability parity, commit/push, open the integration-base PR, and report the API-observed base.

## File ownership expectation

- PostgreSQL bundle/native adapter and tests under `internal/connectors/defs/postgres/` and `internal/connectors/native/postgres/`.
- Narrow generic transport request/warehouse artifact additions under `internal/synctransport/` and production wiring under `internal/app/`.
- CLI parsing/help tests under `internal/cli/`, corresponding `docs/cli/**` and `website/**`, and generated outputs only through their generators.
- This phase directory for lifecycle evidence.

## Guardrails

- No dependency changes and no secrets in argv, output, logs, fixtures, state, plans, or tests.
- No raw DSN/SQL/target table/path input. PostgreSQL identifiers come from validated catalog or derived managed-target values.
- Context reaches connection open, snapshot/bootstrap/CDC, staging, target transaction, baseline, read-back, and cleanup. A post-commit baseline write may use `context.WithoutCancel` only as the existing receipt contract specifies.
- Unknown commit is never retried blindly. Already-acknowledged work is replay-safe through the target ledger; unacknowledged work does not advance either checkpoint.
- Capability `write=false` remains public truth; the private transport destination is not a generic Writer.

## Plan validation

The plan is goal-backward: a shipped-binary call chain cannot exist without declaration/factory registration, actual App dispatch, warehouse-reopened Parquet consumption, target construction, and command-level approval. Each is paired with an observable red/green test, and every high-risk boundary has a failure/no-side-effect assertion.
