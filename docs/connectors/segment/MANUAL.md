# pm connectors inspect segment

```text
NAME
  pm connectors inspect segment - Segment connector manual

SYNOPSIS
  pm connectors inspect segment
  pm connectors inspect segment --json
  pm credentials add <name> --connector segment [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Segment workspace, source, and destination metadata through the Segment Public API.

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
  api_token (secret)

ETL STREAMS
  workspaces:
    primary key: id
    fields: id(), name(), slug(), updated_at()
  sources:
    primary key: id
    fields: id(), name(), slug(), updated_at()
  destinations:
    primary key: id
    fields: id(), name(), slug(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Segment Public API read of workspace, source, and destination metadata
  approval: none; read-only, no reverse-ETL writes implemented by legacy
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect segment

  # Inspect as structured JSON
  pm connectors inspect segment --json

AGENT WORKFLOW
  - Run pm connectors inspect segment before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
