# pm connectors inspect postgres

```text
NAME
  pm connectors inspect postgres - PostgreSQL connector manual

SYNOPSIS
  pm connectors inspect postgres
  pm connectors inspect postgres --json
  pm credentials add <name> --connector postgres [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and supports cursor-incremental reads. Read-only source.

ICON
  id: postgresql
  asset: icons/postgresql.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.postgresql.org/docs/current/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: database

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  cdc_publication
  cursor_field
  database
  host
  mode
  port
  read_limit
  schema
  sslmode
  username
  password (secret)

SECURITY
  read risk: low
  write risk: n/a (read-only source)
  approval: none required for read-only sync
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect postgres

  # Inspect as structured JSON
  pm connectors inspect postgres --json

AGENT WORKFLOW
  - Run pm connectors inspect postgres before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
