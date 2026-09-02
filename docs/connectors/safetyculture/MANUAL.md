# pm connectors inspect safetyculture

```text
NAME
  pm connectors inspect safetyculture - SafetyCulture connector manual

SYNOPSIS
  pm connectors inspect safetyculture
  pm connectors inspect safetyculture --json
  pm credentials add <name> --connector safetyculture [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SafetyCulture audits, templates, and users through fixed REST routes.

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
  access_token (secret) (required)

ETL STREAMS
  audits:
    primary key: id
    fields: id(string), modified_at(string), name(string)
  templates:
    primary key: id
    fields: id(string), modified_at(string), name(string)
  users:
    primary key: id
    fields: id(string), modified_at(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: Bounded GET reads use the fixed SafetyCulture origin and declared bearer authentication.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect safetyculture

  # Inspect as structured JSON
  pm connectors inspect safetyculture --json

AGENT WORKFLOW
  - Run pm connectors inspect safetyculture before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
