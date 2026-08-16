```
NAME
  pm etl - run local ETL syncs

SYNOPSIS
  pm etl check --connector <name> [--config key=value] [--json]
  pm etl catalog --connector <name> [--config key=value] [--json]
  pm etl read --connector <name> [--stream stream] [--limit n] [--config key=value] [--json]
  pm etl run --connection <name> --stream <stream> [--batch-size n] [--runtime] [--json]
  pm etl status <run-id> [--json]
  pm etl transport github-issue-label plan --connection <name> [--json]
  pm etl transport github-issue-label preview <plan-id> [--json]
  pm etl run --connection <name> --stream issues --batch-size 1 --approval-plan <plan-id> [--approval-token-stdin] --confirm destructive [--json]
  pm etl transport postgres-managed-target plan --connection <name> --stream <stream> [--authorization-lifetime <24h..48h>] [--json]
  pm etl transport postgres-managed-target preview <plan-id> [--json]
  pm etl transport github-issue-label cleanup plan --connection <name> --forward-plan <plan-id> [--json]
  pm etl transport github-issue-label cleanup run <plan-id> --connection <name> --approval-token-stdin --confirm destructive [--json]

DESCRIPTION
  ETL can directly check, catalog, and read enabled connectors by name. The
  read surface comes from connector definitions: declarative JSON bundles
  interpreted by the connector engine, with hooks or native components where an
  API or protocol needs custom behavior. Use pm connectors inspect <name> to
  see available streams.

  Some catalog slugs remain migration metadata only. Those entries are still
  inspectable through pm connectors inspect, but cannot execute ETL until a
  runnable connector definition or component passes conformance and is enabled.

  ETL runs read records from a configured source connector stream, add
  Polymetrics metadata fields, and write records to the destination connector.
  The warehouse destination uses an appendable JSONL write-ahead log and rebuilds
  each final table as a single Parquet file.

  ETL and reverse ETL are separate first-class connector surfaces: ETL reads
  streams, while pm reverse executes connector write actions where the upstream
  API supports mutations.

  ETL writes destination records in bounded batches. Use --batch-size for large
  paginated streams when you want tighter memory bounds.

  With --runtime, ETL also requires healthy PostgreSQL, DragonflyDB, and Temporal
  endpoints. It acquires a Dragonfly lease and appends a PostgreSQL run-ledger
  record after the local ETL completes.

CLOSED GITHUB TRANSPORT
  The github-issue-label transport is a fixed GitHub issue-to-label walking
  slice. A saved connection owns the repository, source issue, target issue,
  label, action, and credential configuration; the command accepts none of
  those provider details directly.

  Create a closed plan, preview it in human output to obtain an ephemeral
  approval token, then pass that token only as one bounded stdin line to:

    pm etl run --connection <name> --stream issues --batch-size 1 \
      --approval-plan <plan-id> --approval-token-stdin --confirm destructive

  The run keeps the source -> durable warehouse -> reopen -> typed GitHub
  mutation and durable acknowledgement -> independent read-back -> checkpoint
  order. After the first approved non-additive run, later runs with the exact
  same plan scope use --approval-plan and --confirm destructive without a new
  --approval-token-stdin. Changed, expired, or revoked scope is refused before
  a provider write. Cleanup is a separate typed remove-label plan, preview, and
  one-time approval. A declared GitHub missing-label DELETE is a successful
  cleanup; replaying approval is refused.

  Source selection and independent read-back each inspect only the first GitHub
  issues page. The transport fails instead of requesting another page when the
  configured source or target issue is not there.

  Approval tokens are never accepted in argv, environment variables, files,
  JSON output, or persisted project state. Run pm etl transport for the exact
  closed lifecycle and its stdin-only token rule.

CLOSED POSTGRESQL MANAGED-TARGET TRANSPORT
  postgres-managed-target is a fixed definition-selected source-to-PostgreSQL
  route, not a generic SQL writer. The saved connection binds both credentials,
  the source stream and sealed schema, primary key, mode, and immutable managed
  destination identity. PostgreSQL sources use their typed relation catalog;
  declared API sources use their authoritative JSON stream schema.

  Create and preview a plan, then pass its one-time token only through stdin to
  the ordinary ETL run:

    pm etl transport postgres-managed-target plan \
      --connection <name> --stream <stream> [--authorization-lifetime <24h..48h>] --json
    pm etl transport postgres-managed-target preview <plan-id>
    pm etl run --connection <name> --stream <stream> --batch-size 1000 \
      --approval-plan <plan-id> --approval-token-stdin --confirm destructive

  The registered source stages bounded pages in the connection-owned warehouse.
  The registered destination reopens the sealed Parquet, derives an immutable
  workset, applies it through the native managed target, reads the receipt back,
  and advances the source checkpoint only after durable acknowledgement.
  With transport_bootstrap=true on an incremental_upsert PostgreSQL source
  credential, a gap-free snapshot barrier continues into committed pgoutput
  transactions and resumes from the acknowledged LSN after restart.

  Approval binds the live source schema and both credential revisions. The
  one-time preview token creates a durable, revocable authorization with a
  24h default lifetime (configurable from 24h through 48h at plan time).
  Each provider page fetch and staged PostgreSQL apply has its own bounded
  deadline, so a stalled unit stops without expiring the overall authority.
  Stale, replayed, authentication-refused, and permission-refused runs stop
  before a checkpoint advance. The public PostgreSQL connector remains
  write=false and this route accepts no raw SQL or target identifiers.

DIRECT CONNECTOR COMMANDS
  check
    Calls the connector check operation and returns status=ok on success.

  catalog
    Calls the connector catalog/discover operation and prints available streams.

  read
    Reads fixture-backed or live records from a connector stream with a hard
    output limit. Use --json for stable agent output.

SOURCE STREAMS
  sample.customers
    Deterministic customer fixture stream. Primary key: id. Cursor: updated_at.

  sample.events
    Deterministic event fixture stream. Primary key: id. Cursor: occurred_at.

  file.file
    Local JSONL or CSV file stream. Configure path and optionally stream.

  github.issues
    Repository issues excluding pull requests. Primary key: node_id. Cursor:
    updated_at. Supports public, token, and github_app auth.

  github.pull_requests
    Repository pull requests. Primary key: node_id. Cursor: updated_at.
    Supports public, token, and github_app auth.

DESTINATIONS
  warehouse
    Local Parquet warehouse tables. Supports append, overwrite, append_dedup, and
    overwrite_dedup destination behavior through ETL sync modes.

SYNC MODES
  full_refresh_append
    Reads every source record and appends to the write-ahead log, then rebuilds
    the final Parquet table from it. Duplicates across runs are expected.

  full_refresh_overwrite
    Replaces the write-ahead log with this run's records, then atomically
    replaces the final Parquet table only after the run succeeds.

  full_refresh_overwrite_deduped
    Compatibility name for typed full_overwrite admission. pm refuses before
    source I/O until a matching transport is admitted.

  incremental_append
    Reads records at or after the saved cursor and appends accepted records to
    the write-ahead log. Cursor state advances only after successful writes.

  incremental_append_deduped
    Compatibility name for typed incremental_dedupe admission. pm refuses
    before source I/O until a matching transport is admitted.

  Incremental modes and deduped compatibility names require --cursor. Deduped
  modes require --primary-key. Static connector manifests advertise the full
  deduped compatibility name only with both fields, and incremental modes only
  with a declared incremental executor.

POLLING-WATERMARK LIMITS
  polling_watermark is a bounded keyset scan, not CDC or a generic database
  query. The runtime evaluates every mode against the declared source ordering,
  discovered object, destination binding, registered executors, and conformance
  evidence before source I/O. A durable checkpoint records the watermark and
  unique tie-breaker only after downstream acknowledgement, so accepted records
  may replay. Hard deletes are not observable unless the declaration supplies
  a cursor-advancing soft-delete mapping. Incompatible state, source identity
  mismatch, snapshot expiry, and retention failure require explicit rebootstrap;
  pm never converts them into an automatic full scan.

SECURITY
  ETL resolves credentials in memory and stores only credential references.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
