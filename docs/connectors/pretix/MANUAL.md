# pm connectors inspect pretix

```text
NAME
  pm connectors inspect pretix - Pretix connector manual

SYNOPSIS
  pm connectors inspect pretix
  pm connectors inspect pretix --json
  pm credentials add <name> --connector pretix [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads pretix organizers, events, items, and orders through the pretix REST API.

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
  event
  organizer
  api_token (secret) (required)

ETL STREAMS
  organizers:
    primary key: id
    fields: id(string), name(object), slug(string)
  events:
    primary key: id
    fields: id(string), name(object), slug(string), updated_at(string)
  items:
    primary key: id
    fields: code(string), id(integer), name(object), slug(string)
  orders:
    primary key: id
    fields: code(string), id(string), name(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external pretix API read of organizer, event, item, and order data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pretix

  # Inspect as structured JSON
  pm connectors inspect pretix --json

AGENT WORKFLOW
  - Run pm connectors inspect pretix before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
