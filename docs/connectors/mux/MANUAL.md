# pm connectors inspect mux

```text
NAME
  pm connectors inspect mux - Mux connector manual

SYNOPSIS
  pm connectors inspect mux
  pm connectors inspect mux --json
  pm credentials add <name> --connector mux [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mux Video assets, live streams, direct uploads, and system signing keys through the Mux REST API using HTTP Basic authentication.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  username
  password (secret)

ETL STREAMS
  assets:
    primary key: id
    cursor: created_at
    fields: created_at(), duration(), encoding_tier(), id(), master_access(), max_resolution_tier(), mp4_support(), status(), test()
  live_streams:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), latency_mode(), max_continuous_duration(), reconnect_window(), status(), stream_key(), test()
  uploads:
    primary key: id
    fields: asset_id(), cors_origin(), id(), status(), test(), timeout(), url()
  signing_keys:
    primary key: id
    cursor: created_at
    fields: created_at(), id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Mux API read of video asset, live stream, upload, and signing key data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mux

  # Inspect as structured JSON
  pm connectors inspect mux --json

AGENT WORKFLOW
  - Run pm connectors inspect mux before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
