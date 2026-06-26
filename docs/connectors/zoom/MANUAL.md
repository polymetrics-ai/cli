# pm connectors inspect zoom

```text
NAME
  pm connectors inspect zoom - Zoom connector manual

SYNOPSIS
  pm connectors inspect zoom
  pm connectors inspect zoom --json
  pm credentials add <name> --connector zoom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoom users, meetings, and webinars through the Zoom REST API.

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
  pm connectors inspect zoom

  # Inspect as structured JSON
  pm connectors inspect zoom --json

AGENT WORKFLOW
  - Run pm connectors inspect zoom before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
