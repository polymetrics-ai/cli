# pm connectors inspect drift

```text
NAME
  pm connectors inspect drift - Drift connector manual

SYNOPSIS
  pm connectors inspect drift
  pm connectors inspect drift --json
  pm credentials add <name> --connector drift [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Drift users, accounts, conversations, and contacts through the Drift REST API.

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
  pm connectors inspect drift

  # Inspect as structured JSON
  pm connectors inspect drift --json

AGENT WORKFLOW
  - Run pm connectors inspect drift before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
