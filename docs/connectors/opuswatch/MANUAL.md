# pm connectors inspect opuswatch

```text
NAME
  pm connectors inspect opuswatch - OPUSWatch connector manual

SYNOPSIS
  pm connectors inspect opuswatch
  pm connectors inspect opuswatch --json
  pm credentials add <name> --connector opuswatch [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads OPUSWatch monitors, incidents, and checks.

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
  page_size
  api_key (secret)

ETL STREAMS
  monitors:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), message(), name(), status(), updated_at()
  incidents:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), message(), name(), status(), updated_at()
  checks:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), message(), name(), status(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external OPUSWatch API read of monitor, incident, and check status data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect opuswatch

  # Inspect as structured JSON
  pm connectors inspect opuswatch --json

AGENT WORKFLOW
  - Run pm connectors inspect opuswatch before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
