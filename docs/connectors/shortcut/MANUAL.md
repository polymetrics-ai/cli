# pm connectors inspect shortcut

```text
NAME
  pm connectors inspect shortcut - Shortcut connector manual

SYNOPSIS
  pm connectors inspect shortcut
  pm connectors inspect shortcut --json
  pm credentials add <name> --connector shortcut [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Shortcut stories, epics, projects, and iterations through the Shortcut REST API.

ICON
  asset: icons/shortcut.svg
  source: official
  review_status: official_verified
  review_url: https://developer.shortcut.com/api/rest/v3

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  page_size
  api_token (secret)

ETL STREAMS
  stories:
    primary key: id
    cursor: updated_at
    fields: id(), name(), state(), updated_at()
  epics:
    primary key: id
    cursor: updated_at
    fields: id(), name(), state(), updated_at()
  projects:
    primary key: id
    cursor: updated_at
    fields: id(), name(), state(), updated_at()
  iterations:
    primary key: id
    cursor: updated_at
    fields: id(), name(), state(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Shortcut API read of story, epic, project, and iteration data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect shortcut

  # Inspect as structured JSON
  pm connectors inspect shortcut --json

AGENT WORKFLOW
  - Run pm connectors inspect shortcut before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
