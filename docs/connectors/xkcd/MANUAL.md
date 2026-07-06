# pm connectors inspect xkcd

```text
NAME
  pm connectors inspect xkcd - XKCD connector manual

SYNOPSIS
  pm connectors inspect xkcd
  pm connectors inspect xkcd --json
  pm credentials add <name> --connector xkcd [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public XKCD comic metadata from the JSON API. Read-only.

ICON
  asset: icons/xkcd.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://xkcd.com/json.html

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  comic_number

ETL STREAMS
  latest:
    primary key: num
    fields: alt(), day(), img(), link(), month(), news(), num(), safe_title(), title(), transcript(), year()
  comic:
    primary key: num
    fields: alt(), day(), img(), link(), month(), news(), num(), safe_title(), title(), transcript(), year()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: public XKCD comic metadata read, no credentials involved
  approval: none; read-only public API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect xkcd

  # Inspect as structured JSON
  pm connectors inspect xkcd --json

AGENT WORKFLOW
  - Run pm connectors inspect xkcd before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
