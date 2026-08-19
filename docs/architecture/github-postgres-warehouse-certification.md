# GitHub and PostgreSQL connector release certification through DuckDB/Parquet

## Outcome

Implement one shared sync engine around Parquet and embedded DuckDB, with GitHub and PostgreSQL supplying transport-specific adapters. Do not build independent GitHub and PostgreSQL synchronization pipelines.

The required independently certified legs are:

1. GitHub API source to DuckDB/Parquet warehouse.
2. DuckDB/Parquet warehouse to GitHub API destination.
3. PostgreSQL source to DuckDB/Parquet warehouse.
4. DuckDB/Parquet warehouse to managed PostgreSQL destination.

Once these legs share one contract, the system can compose GitHub to warehouse to GitHub, PostgreSQL to warehouse to PostgreSQL, GitHub to warehouse to PostgreSQL, and PostgreSQL to warehouse to GitHub. At least one cross-connector path must be certified to prove that the warehouse is the mediator rather than connector-specific glue.

## Target architecture

```text
GitHub API source -- Connector.Read ----+
                                        +-- Immutable captured batch
PostgreSQL source -- native source -----+   records + schema + checkpoint
                     (definition-selected)
                                        |
                                        v
                 Parquet/DuckDB sync executor
            WAL -> validate -> materialize -> receipt
                             |
                  +----------+----------+
                  |                     |
             Warehouse read       Warehouse query
                  |                     |
                  v                     v
            Validation/API       Reverse-ETL plan
                                   + approval hash
                                          |
                          +---------------+---------------+
                          v                               v
                   GitHub Connector.Write        PostgreSQL Connector.Write
```

DuckDB is the common materialization and query engine. The durable warehouse is a connection-owned JSONL WAL plus one derived Parquet file for each table.

PostgreSQL's native source selection is declaration-led rather than App-composed. Its exact
snapshot streams, modes, and limits are owned by the
[PostgreSQL bundle docs](../../internal/connectors/defs/postgres/docs.md).

## Responsibility boundaries

### Connector adapters

Connector adapters own authentication and connections, API pagination or database cursors, rate limits and typed transport failures, source record reads, destination record writes, destination readback, and truthful declarations of supported modes and actions.

### Shared sync engine

The shared engine owns canonical mode admission, checkpoint and resume behavior, immutable workset identity, dedupe and history semantics, approval and plan hashing, replay and idempotency, receipts, and the execution used by direct commands, ETL, reverse ETL, flows, and schedules.

### Warehouse

The warehouse owns the connection-scoped JSONL WAL, fsync before acknowledgement, DuckDB transformation and validation, one derived Parquet table file, atomic file replacement, row counts and content hashes, durable receipts, and readback for outbound plans.

GitHub and PostgreSQL must not define their own meanings for overwrite, dedupe, upsert, or history.

## Connector-to-warehouse transaction

1. Read a bounded source batch.
2. Produce a candidate source checkpoint without persisting it.
3. Append the batch to the WAL and record its immutable workset identity.
4. Fsync the WAL.
5. Apply the selected canonical sync mode using DuckDB.
6. Materialize the resulting single Parquet table file and atomically replace the prior file.
7. Write a warehouse receipt containing the connector, stream, sync mode, run and workset IDs, input/output counts, schema hash, content hash, and candidate checkpoint.
8. Read the Parquet result back through DuckDB and validate it.
9. Advance the source checkpoint only after the receipt and readback succeed.

## Warehouse-to-connector transaction

1. Pin an immutable warehouse workset and reopen its connection-owned Parquet table.
2. Query it through DuckDB.
3. Normalize the exact destination operations.
4. Hash the complete plan.
5. Obtain a plan-bound approval or scoped scheduled grant.
6. Execute `Connector.Write`.
7. Record per-record and batch destination receipts.
8. Read the destination back.
9. Compare keys, values, counts, and hashes.
10. Mark the workset delivered only after the receipt and readback succeed.

This ordering ensures that checkpoints never advance ahead of durable data or destination acknowledgement.

## Canonical sync modes

`full_refresh` is a source semantic used by the two full modes, not a separate canonical mode.

| Canonical mode | Shared warehouse behavior |
| --- | --- |
| `full_overwrite` | Build and validate one staged Parquet file, then atomically replace the active table file. |
| `full_append` | Append the complete new snapshot to the WAL, rebuild the active table file from that WAL, then atomically replace it. |
| `incremental_append` | Append only records after the exact checkpoint and make replay idempotent through workset identity. |
| `incremental_upsert` | Partition by primary key and retain the latest cursor/version, including tombstone behavior. |
| `incremental_dedupe` | Fold input to one current non-deleted record per primary key. |
| `incremental_dedupe_history` | Produce SCD2 history with `_valid_from`, `_valid_to`, `_is_current`, and stable source-version identity. |
| `change_capture` | Preserve whole database transactions and commit the warehouse receipt before acknowledging the source LSN. |

Deterministic dedupe requires a primary key, source cursor or version, stable tie-breaker, deletion marker, and ingestion sequence as the final fallback.

## GitHub implementation

### GitHub to warehouse

Reuse the declarative GitHub `Connector.Read`, but route canonical modes through the shared executor instead of the legacy-mode bypass. Prove deterministic pagination, stable record identity, cursor semantics where supported, typed rate-limit/retry failures, schema/content hashes, and exact full, incremental, upsert, dedupe, and history warehouse semantics.

### Warehouse to GitHub

Reuse `PlanReverseETL` and `RunReverseETL`; do not use an arbitrary-URL HTTP action runner. Each write action, such as `create_issue`, `update_issue`, `add_comment`, or `set_labels`, must be closed and typed with an exact payload schema, destination identity mapping, preview, plan hash, approval, idempotency or reconciliation behavior, response receipt, and GitHub readback.

Flows and schedules must invoke this same connector writer. A schedule stores a scoped grant reference rather than a reusable raw approval token.

## PostgreSQL implementation

### PostgreSQL to warehouse

The current PostgreSQL source leg is a definition-selected bounded snapshot. Its exact streams,
modes, and checkpoint limits are owned by the
[PostgreSQL bundle docs](../../internal/connectors/defs/postgres/docs.md). Polling is a separate,
future transport and is not a fallback for that snapshot.

Polling resume must use an exact composite position `(cursor_value, primary_key)`. For the architecture's composite contract, bind the prior cursor value as pgx argument 1 (`$1`) and the stable primary-key tie breaker as argument 2 (`$2`); the predicate and ordering must handle equal cursor values:

```sql
WHERE cursor > $1
   OR (cursor = $1 AND primary_key > $2)
ORDER BY cursor, primary_key
```

Any future polling or CDC source must prove a consistent snapshot boundary, schema drift handling,
bounded reads, exact restart, transaction-preserving CDC, and a whole-transaction warehouse receipt
before LSN acknowledgement.

### Warehouse to PostgreSQL

Implement a managed-target driver with strict ownership boundaries. It must pin the DuckDB workset, open a PostgreSQL delivery session, load an owned staging table, validate row counts and hashes, apply append/upsert/overwrite/dedupe in one transaction, store a remote receipt keyed by workset and plan hash, commit, read the target back, and only then advance the delivery checkpoint. Replaying the same workset must return the existing receipt or reproduce the same result without duplicate rows.

## Mode applicability

Certification requires every applicable mode and a typed evidence-bearing outcome for unsupported or non-applicable cells. It must not force unsafe symmetry or infer exact execution from broad `read` or `write` capability flags.

GitHub source to warehouse can support all six non-CDC warehouse modes. Exact PostgreSQL source
availability is owned by the [PostgreSQL bundle docs](../../internal/connectors/defs/postgres/docs.md);
its broad `read` capability does not promote polling, incremental modes, or source CDC. PostgreSQL
destination naturally supports overwrite, append, upsert, and dedupe. GitHub full overwrite requires
an explicit managed-resource ownership and deletion/archive contract. `change_capture` is a source
mode rather than a destination write mode.

The current PostgreSQL #3972 contract records destination-side `incremental_dedupe_history` as typed non-executable. Warehouse-side SCD2 history should be implemented for both sources now. Expanding PostgreSQL destination history requires an explicit managed-history-table contract rather than a false certification claim.

## Certification model

Each result is keyed by connector, connector kind, direction, source or destination role, exact sync mode, workflow (`direct`, `flow`, or `schedule`), environment (`fixture`, provider double/container, or live), and tested commit SHA.

Valid outcomes are `passed`, `failed`, `not_applicable`, `supported_but_untested`, and typed `unsupported`. `implemented` comes from exact executor registration and successful admission, never from a broad capability declaration.

Every applicable passing cell proves actual connector execution, DuckDB/Parquet mediation, warehouse receipt, row and content-hash readback, checkpoint correctness, destination receipt/readback, restart/replay/idempotency, failure boundary, cleanup, and exact build SHA.

## PM release acceptance matrix

The integrated current-SHA binary must execute:

- direct read;
- direct write;
- ETL;
- reverse ETL;
- authored flow execution;
- an actually fired schedule;
- bounded read-only PostgreSQL database query.

The primary gate is the real executable test, not documentation or declaration coverage. A passing fixture cannot substitute for a required provider-double, container, or live cell.

## Issue-first delivery

Umbrella issue #4015 owns one combined draft integration PR to `main`. GitHub parent #3988 and PostgreSQL parent #3972 each own a parent branch and draft PR into the combined branch. Their existing sub-issues retain one scoped implementation PR each. Shared dependencies remain under #3862, especially #3864 and #3866.

Existing #3993 and #4014 work must be recovered and reused rather than discarded or duplicated. Child PRs merge in dependency order into their connector parent; both connector parents then merge into the combined integration branch. The final current-SHA integration PR remains draft until the complete acceptance matrix and full parent checks are green. Merge to `main` requires explicit captain authorization.

## Bounded validation

Use no-mistakes for the applicable child gates and final integrated parent gate, without `--yes`. Limit each PR to five correction or validation cycles. Prioritize executable test failures over non-blocking polish. If five cycles are exhausted while a real test is still red, preserve the branch and evidence and escalate the concrete blocker rather than continuing review churn.
