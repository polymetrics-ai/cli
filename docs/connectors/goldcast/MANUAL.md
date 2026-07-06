# pm connectors inspect goldcast

```text
NAME
  pm connectors inspect goldcast - Goldcast connector manual

SYNOPSIS
  pm connectors inspect goldcast
  pm connectors inspect goldcast --json
  pm credentials add <name> --connector goldcast [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Goldcast organizations, events, agenda items, discussion groups, and tracks through the Goldcast customapi REST API.

ICON
  asset: icons/goldcast.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.goldcast.io/api-docs

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_key (secret)

ETL STREAMS
  organizations:
    primary key: id
    fields: created_at(), domain(), id(), name(), slug()
  events:
    primary key: id
    fields: created_at(), end_time(), id(), organization(), start_time(), status(), timezone(), title()
  agenda_items:
    primary key: id
    fields: description(), end_time(), event(), id(), start_time(), title()
  discussion_groups:
    primary key: id
    fields: capacity(), created_at(), event(), id(), name()
  tracks:
    primary key: id
    fields: color(), event(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Goldcast API read of organization, event, and event-scoped data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect goldcast

  # Inspect as structured JSON
  pm connectors inspect goldcast --json

AGENT WORKFLOW
  - Run pm connectors inspect goldcast before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
