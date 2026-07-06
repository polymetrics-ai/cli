# pm connectors inspect oveit

```text
NAME
  pm connectors inspect oveit - Oveit connector manual

SYNOPSIS
  pm connectors inspect oveit
  pm connectors inspect oveit --json
  pm credentials add <name> --connector oveit [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Oveit events, orders, and attendees.

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
  email
  page_size
  password (secret)

ETL STREAMS
  events:
    primary key: id
    fields: created_at(), email(), id(), name(), starts_at(), status(), total()
  orders:
    primary key: id
    fields: created_at(), email(), id(), name(), starts_at(), status(), total()
  attendees:
    primary key: id
    fields: created_at(), email(), id(), name(), starts_at(), status(), total()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Oveit API read of event, order, and attendee data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect oveit

  # Inspect as structured JSON
  pm connectors inspect oveit --json

AGENT WORKFLOW
  - Run pm connectors inspect oveit before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
