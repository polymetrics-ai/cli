# pm connectors inspect pendo

```text
NAME
  pm connectors inspect pendo - Pendo connector manual

SYNOPSIS
  pm connectors inspect pendo
  pm connectors inspect pendo --json
  pm credentials add <name> --connector pendo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pendo visitors, accounts, pages, and features through REST list endpoints.

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
  pm connectors inspect pendo

  # Inspect as structured JSON
  pm connectors inspect pendo --json

AGENT WORKFLOW
  - Run pm connectors inspect pendo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
