# pm connectors inspect granola

```text
NAME
  pm connectors inspect granola - Granola connector manual

SYNOPSIS
  pm connectors inspect granola
  pm connectors inspect granola --json
  pm credentials add <name> --connector granola [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Granola meeting notes metadata and full note detail (summary, owner, attendees, calendar event) through the Granola public API (read-only).

ICON
  asset: icons/source-granola.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.granola.ai/introduction

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  start_date
  api_key (secret)

ETL STREAMS
  notes:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), object(), owner_email(), owner_name(), title(), updated_at()
  detailed_notes:
    primary key: id
    cursor: created_at
    fields: attendees(), calendar_event(), created_at(), folders(), id(), object(), owner_email(), owner_name(), summary(), title(), transcript(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Granola API read of meeting notes metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect granola

  # Inspect as structured JSON
  pm connectors inspect granola --json

AGENT WORKFLOW
  - Run pm connectors inspect granola before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
