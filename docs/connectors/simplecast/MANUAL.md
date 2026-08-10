# pm connectors inspect simplecast

```text
NAME
  pm connectors inspect simplecast - Simplecast connector manual

SYNOPSIS
  pm connectors inspect simplecast
  pm connectors inspect simplecast --json
  pm credentials add <name> --connector simplecast [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Simplecast podcasts and episodes through the Simplecast REST API.

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
  access_token (secret) (required)

ETL STREAMS
  podcasts:
    primary key: id
    cursor: updated_at
    fields: id(string), status(string), title(string), updated_at(string)
  episodes:
    primary key: id
    cursor: updated_at
    fields: id(string), status(string), title(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Simplecast API read of podcast and episode data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect simplecast

  # Inspect as structured JSON
  pm connectors inspect simplecast --json

AGENT WORKFLOW
  - Run pm connectors inspect simplecast before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
