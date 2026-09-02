# pm connectors inspect copper

```text
NAME
  pm connectors inspect copper - Copper connector manual

SYNOPSIS
  pm connectors inspect copper
  pm connectors inspect copper --json
  pm credentials add <name> --connector copper [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Copper CRM records through fixed typed search routes.

ICON
  id: copper
  asset: icons/copper.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.copper.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  user_email (required)
  api_key (secret) (required)

ETL STREAMS
  people:
    primary key: id
    cursor: date_modified
    fields: date_modified(integer), id(integer)
  companies:
    primary key: id
    cursor: date_modified
    fields: date_modified(integer), id(integer)
  opportunities:
    primary key: id
    cursor: date_modified
    fields: date_modified(integer), id(integer)
  leads:
    primary key: id
    cursor: date_modified
    fields: date_modified(integer), id(integer)
  tasks:
    primary key: id
    cursor: date_modified
    fields: date_modified(integer), id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: Bounded Copper search requests use fixed API routes and declared three-header authentication.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect copper

  # Inspect as structured JSON
  pm connectors inspect copper --json

AGENT WORKFLOW
  - Run pm connectors inspect copper before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
