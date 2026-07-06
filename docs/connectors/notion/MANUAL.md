# pm connectors inspect notion

```text
NAME
  pm connectors inspect notion - Notion connector manual

SYNOPSIS
  pm connectors inspect notion
  pm connectors inspect notion --json
  pm credentials add <name> --connector notion [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Notion databases, pages, and users through the Notion REST API. Read-only.

ICON
  asset: icons/notion.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.notion.com/reference/changes-by-version

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  page_size
  token (secret)

ETL STREAMS
  databases:
    primary key: id
    cursor: last_edited_time
    fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), title(), url()
  pages:
    primary key: id
    cursor: last_edited_time
    fields: archived(), created_time(), id(), in_trash(), last_edited_time(), object(), parent(), properties(), url()
  users:
    primary key: id
    fields: avatar_url(), bot(), id(), name(), object(), person(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Notion API read of workspace databases/pages/users
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect notion

  # Inspect as structured JSON
  pm connectors inspect notion --json

AGENT WORKFLOW
  - Run pm connectors inspect notion before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
