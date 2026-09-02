# pm connectors inspect rootly

```text
NAME
  pm connectors inspect rootly - Rootly connector manual

SYNOPSIS
  pm connectors inspect rootly
  pm connectors inspect rootly --json
  pm credentials add <name> --connector rootly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Rootly incidents, services, and users through fixed JSON:API routes.

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
  start_date (required)
  api_key (secret) (required)

ETL STREAMS
  incidents:
    primary key: id
    fields: id(string), status(string), title(string)
  services:
    primary key: id
    fields: id(string), status(string), title(string)
  users:
    primary key: id
    fields: id(string), status(string), title(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: Bounded Rootly JSON:API reads use the fixed provider origin and declared bearer authentication.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect rootly

  # Inspect as structured JSON
  pm connectors inspect rootly --json

AGENT WORKFLOW
  - Run pm connectors inspect rootly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
