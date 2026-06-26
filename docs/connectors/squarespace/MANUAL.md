# pm connectors inspect squarespace

```text
NAME
  pm connectors inspect squarespace - Squarespace connector manual

SYNOPSIS
  pm connectors inspect squarespace
  pm connectors inspect squarespace --json
  pm credentials add <name> --connector squarespace [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Squarespace orders, products, inventory items, and profiles through the Squarespace API.

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
  pm connectors inspect squarespace

  # Inspect as structured JSON
  pm connectors inspect squarespace --json

AGENT WORKFLOW
  - Run pm connectors inspect squarespace before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
