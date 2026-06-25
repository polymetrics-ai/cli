# pm connectors inspect postgres

```text
NAME
  pm connectors inspect postgres - PostgreSQL connector manual

SYNOPSIS
  pm connectors inspect postgres
  pm connectors inspect postgres --json
  pm credentials add <name> --connector postgres [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and supports cursor-incremental reads on a configurable cursor column. Read-only source; CDC is a documented stub pending the gated pglogrepl dependency.

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: database

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  No connector-specific config fields.

SECURITY
  read risk: connector-specific
  write risk: connector-specific
  approval: external mutations require preview and approval
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
