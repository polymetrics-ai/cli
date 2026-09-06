# pm connectors inspect dixa

```text
NAME
  pm connectors inspect dixa - Dixa connector manual

SYNOPSIS
  pm connectors inspect dixa
  pm connectors inspect dixa --json
  pm credentials add <name> --connector dixa [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Dixa conversation export records through fixed bearer-authenticated export routes.

ICON
  id: dixa
  asset: icons/dixa.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.dixa.io/openapi/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  updated_after
  updated_before
  api_token (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: updated_at
    fields: id(integer), updated_at(integer)
  conversation_queue:
    primary key: id
    cursor: updated_at
    fields: id(integer), updated_at(integer)
  conversation_rating:
    primary key: id
    cursor: updated_at
    fields: id(integer), updated_at(integer)
  conversation_assignment:
    primary key: id
    cursor: updated_at
    fields: id(integer), updated_at(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: Bounded Dixa conversation export reads use a fixed provider origin and declared bearer authentication.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect dixa

  # Inspect as structured JSON
  pm connectors inspect dixa --json

AGENT WORKFLOW
  - Run pm connectors inspect dixa before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
