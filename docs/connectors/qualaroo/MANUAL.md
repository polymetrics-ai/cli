# pm connectors inspect qualaroo

```text
NAME
  pm connectors inspect qualaroo - Qualaroo connector manual

SYNOPSIS
  pm connectors inspect qualaroo
  pm connectors inspect qualaroo --json
  pm credentials add <name> --connector qualaroo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Qualaroo nudges and response records through the Qualaroo API. Read-only.

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
  pm connectors inspect qualaroo

  # Inspect as structured JSON
  pm connectors inspect qualaroo --json

AGENT WORKFLOW
  - Run pm connectors inspect qualaroo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
