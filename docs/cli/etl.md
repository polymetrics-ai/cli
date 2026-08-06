```
NAME
  pm etl - run local ETL syncs

SYNOPSIS
  pm etl check --connector <name> [--config key=value] [--json]
  pm etl catalog --connector <name> [--config key=value] [--json]
  pm etl read --connector <name> [--stream stream] [--limit n] [--config key=value] [--json]
  pm etl run --connection <name> --stream <stream> [--batch-size n] [--runtime] [--json]
  pm etl status <run-id> [--json]

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
  The MVP warehouse destination stores tables as JSONL files.

  ETL and reverse ETL are separate first-class connector surfaces: ETL reads
  streams, while pm reverse executes connector write actions where the upstream
  API supports mutations.

  ETL writes destination records in bounded batches. Use --batch-size for large
  paginated streams when you want tighter memory bounds.

  With --runtime, ETL also requires healthy PostgreSQL, DragonflyDB, and Temporal
  endpoints. It acquires a Dragonfly lease and appends a PostgreSQL run-ledger
  record after the local ETL completes.

RATE LIMIT OUTPUT
  Every emitted ETL run result and status for a stored run includes a bounded
  rate-limit summary at run.rate_limit. Preflight errors that prevent creation
  of a run use the ordinary Error output and do not invent a run summary.

  Human output labels local_pacing_wait, provider_429_observed,
  provider_429_honored, provider_429_wait, and request_latency separately so a
  slow run does not look like one undifferentiated delay. JSON uses
  pacing_wait_ms, provider_429_observed, provider_429_honored,
  provider_wait_ms, request_latency_ms, and request_count.

  A declared connector reports selected policy IDs, subject kinds, structural
  selection reasons, and typed provider budget facts when available. Declaration
  is declared, unknown, not_applicable, or undeclared. An undeclared connector
  has no rate_limits.json; it is not a claim that provider traffic is unlimited.
  Policies are coalesced rather than emitted per request; policies_omitted
  reports any selected policies omitted from the bounded summary.

  Rate-limit output never includes credentials, token-derived or runtime subject
  values, raw bindings, scope values, configuration, request headers, URLs,
  bodies, or credential revisions.

  Failed pm etl run and pm etl status results emit a failed run ID and the same
  bounded summary. Their JSON failed-run carrier contains only id, status, and
  rate_limit; it omits error text and checkpoints.

  Completed-run JSON contains only run ID and type, status, counters, timing,
  and rate-limit facts. It omits connections, streams, error text, and
  checkpoints.

  If --runtime recording fails after ETL completes, pm still emits the completed
  run and summary, then exits non-zero. It reports runtime_recorded=false and
  runtime_recording=failed so incomplete runtime recording is explicit.

  If completed-run state persistence fails, pm still emits the completed run and
  summary, then exits non-zero with persistence=failed.

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
    Local JSONL warehouse tables. Supports append, overwrite, append_dedup, and
    overwrite_dedup destination behavior through ETL sync modes.

SYNC MODES
  full_refresh_append
    Reads every source record and appends to the final JSONL table. Duplicates
    across runs are expected.

  full_refresh_overwrite
    Reads every source record into a temp final file, then atomically replaces
    the final JSONL table only after the run succeeds.

  full_refresh_overwrite_deduped
    Reads every source record, writes current-generation raw JSONL, dedupes by
    primary key and cursor, then atomically replaces the final JSONL table.

  incremental_append
    Reads records at or after the saved cursor and appends accepted records.
    Cursor state advances only after successful writes.

  incremental_append_deduped
    Appends accepted records to raw JSONL history and materializes a final JSONL
    table with one latest row per primary key. Delete/tombstone records remove
    the row from final output.

SECURITY
  ETL resolves credentials in memory and stores only credential references.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
