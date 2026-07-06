# pm connectors inspect recreation

```text
NAME
  pm connectors inspect recreation - Recreation.gov connector manual

SYNOPSIS
  pm connectors inspect recreation
  pm connectors inspect recreation --json
  pm credentials add <name> --connector recreation [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Recreation.gov RIDB facilities, campsites, activities, organizations, and recreation areas through the RIDB REST API.

ICON
  asset: icons/recreation.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  facilities:
    primary key: id
    cursor: updated_at
    fields: id(), name(), type(), updated_at()
  campsites:
    primary key: id
    cursor: updated_at
    fields: id(), name(), type(), updated_at()
  activities:
    primary key: id
    fields: id(), name()
  organizations:
    primary key: id
    fields: id(), name()
  recareas:
    primary key: id
    fields: id(), name(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Recreation.gov RIDB API read of public facility, campsite, activity, organization, and recreation-area data
  approval: none; read-only public-data API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect recreation

  # Inspect as structured JSON
  pm connectors inspect recreation --json

AGENT WORKFLOW
  - Run pm connectors inspect recreation before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
