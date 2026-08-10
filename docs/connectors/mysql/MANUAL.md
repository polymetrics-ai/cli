# pm connectors inspect mysql

```text
NAME
  pm connectors inspect mysql - MySQL connector manual

SYNOPSIS
  pm connectors inspect mysql
  pm connectors inspect mysql --json
  pm credentials add <name> --connector mysql [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Native MySQL source connector for wire-protocol checks, dynamic schemas, and bounded reads. Read-only source.

ICON
  id: mysql
  asset: icons/mysql.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://dev.mysql.com/doc/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: database

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  cursor_field
  database (required)
  host (required)
  page_size
  port
  read_limit
  sslmode
  sslrootcert
  sslservername
  username (required)
  password (secret)

SECURITY
  read risk: read-only MySQL wire-protocol queries against the configured database
  write risk: n/a (read-only source)
  approval: none required for read-only sync
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mysql

  # Inspect as structured JSON
  pm connectors inspect mysql --json

AGENT WORKFLOW
  - Run pm connectors inspect mysql before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
