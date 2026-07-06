# pm connectors inspect zoom

```text
NAME
  pm connectors inspect zoom - Zoom connector manual

SYNOPSIS
  pm connectors inspect zoom
  pm connectors inspect zoom --json
  pm credentials add <name> --connector zoom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoom users, meetings, and webinars through the Zoom REST API.

ICON
  asset: icons/zoom.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.zoom.us/docs/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  user_id
  access_token (secret)

ETL STREAMS
  users:
    primary key: id
    fields: email(), id(), name(), updated_at()
  meetings:
    primary key: id
    fields: email(), id(), name(), updated_at()
  webinars:
    primary key: id
    fields: email(), id(), name(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Zoom API read of user, meeting, and webinar data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoom

  # Inspect as structured JSON
  pm connectors inspect zoom --json

AGENT WORKFLOW
  - Run pm connectors inspect zoom before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
