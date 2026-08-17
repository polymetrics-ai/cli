# pm connectors inspect shortio

```text
NAME
  pm connectors inspect shortio - Short.io connector manual

SYNOPSIS
  pm connectors inspect shortio
  pm connectors inspect shortio --json
  pm credentials add <name> --connector shortio [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Short.io links and domains through the Short.io REST API.

ICON
  asset: icons/shortio.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.short.io/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  page_size
  api_key (secret)

ETL STREAMS
  links:
    primary key: id
    cursor: updated_at
    fields: id(), name(), path(), title(), updated_at()
  domains:
    primary key: id
    cursor: updated_at
    fields: id(), name(), path(), title(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Short.io API read of link and domain data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect shortio

  # Inspect as structured JSON
  pm connectors inspect shortio --json

AGENT WORKFLOW
  - Run pm connectors inspect shortio before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
