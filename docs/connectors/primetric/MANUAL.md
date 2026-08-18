# pm connectors inspect primetric

```text
NAME
  pm connectors inspect primetric - Primetric connector manual

SYNOPSIS
  pm connectors inspect primetric
  pm connectors inspect primetric --json
  pm credentials add <name> --connector primetric [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Primetric employees, projects, clients, and roles through OAuth-authenticated REST list endpoints.

ICON
  id: primetric
  asset: icons/primetric.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.primetric.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  token_url
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  employees:
    primary key: id
    fields: created_at(string), email(string), first_name(string), id(integer), last_name(string), name(string), updated_at(string)
  projects:
    primary key: id
    fields: created_at(string), id(integer), name(string), updated_at(string)
  clients:
    primary key: id
    fields: created_at(string), id(integer), name(string), updated_at(string)
  roles:
    primary key: id
    fields: created_at(string), id(integer), name(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Primetric API read of employee, project, client, and role data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect primetric

  # Inspect as structured JSON
  pm connectors inspect primetric --json

AGENT WORKFLOW
  - Run pm connectors inspect primetric before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
