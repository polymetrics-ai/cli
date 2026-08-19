# pm connectors inspect productive

```text
NAME
  pm connectors inspect productive - Productive connector manual

SYNOPSIS
  pm connectors inspect productive
  pm connectors inspect productive --json
  pm credentials add <name> --connector productive [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Productive projects, people, companies, and tasks through the Productive JSON:API-style REST API (read-only).

ICON
  id: productive
  asset: icons/productive.svg
  source: official
  review_status: official_verified
  review_url: https://developer.productive.io/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  organization_id (required)
  api_key (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), name(string), type(string), updated_at(string)
  people:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), name(string), type(string), updated_at(string)
  companies:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), name(string), type(string), updated_at(string)
  tasks:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), name(string), type(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Productive API read of projects, people, companies, and tasks
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect productive

  # Inspect as structured JSON
  pm connectors inspect productive --json

AGENT WORKFLOW
  - Run pm connectors inspect productive before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
