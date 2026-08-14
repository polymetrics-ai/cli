# pm connectors inspect productboard

```text
NAME
  pm connectors inspect productboard - Productboard connector manual

SYNOPSIS
  pm connectors inspect productboard
  pm connectors inspect productboard --json
  pm credentials add <name> --connector productboard [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Productboard features, notes, components, and products through the public API.

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
  start_date
  access_token (secret) (required)

ETL STREAMS
  features:
    primary key: id
    fields: created_at(string), id(string), name(string), status(object), title(string), updated_at(string)
  notes:
    primary key: id
    fields: created_at(string), id(string), name(string), status(object), title(string), updated_at(string)
  components:
    primary key: id
    fields: created_at(string), id(string), name(string), status(object), title(string), updated_at(string)
  products:
    primary key: id
    fields: created_at(string), id(string), name(string), status(object), title(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Productboard API read of feature, note, component, and product data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect productboard

  # Inspect as structured JSON
  pm connectors inspect productboard --json

AGENT WORKFLOW
  - Run pm connectors inspect productboard before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
