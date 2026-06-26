# pm connectors inspect huntr

```text
NAME
  pm connectors inspect huntr - Huntr connector manual

SYNOPSIS
  pm connectors inspect huntr
  pm connectors inspect huntr --json
  pm credentials add <name> --connector huntr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Huntr organization members, candidates, activities, notes, and actions through the Huntr REST API.

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

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
  pm connectors inspect huntr

  # Inspect as structured JSON
  pm connectors inspect huntr --json

AGENT WORKFLOW
  - Run pm connectors inspect huntr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
