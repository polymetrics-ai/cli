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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  email (required)
  page_size
  password (secret) (required)

ETL STREAMS
  events:
    primary key: id
    fields: created_at(string), email(string), id(string), name(string), starts_at(string), status(string), total(integer)
  orders:
    primary key: id
    fields: created_at(string), email(string), id(string), name(string), starts_at(string), status(string), total(integer)
  attendees:
    primary key: id
    fields: created_at(string), email(string), id(string), name(string), starts_at(string), status(string), total(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
