# pm connectors inspect freightview

```text
NAME
  pm connectors inspect freightview - Freightview connector manual

SYNOPSIS
  pm connectors inspect freightview
  pm connectors inspect freightview --json
  pm credentials add <name> --connector freightview [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freightview shipments, quotes, and tracking events through the Freightview v2.0 REST API using the client-credentials session-token flow.

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
  pm connectors inspect freightview

  # Inspect as structured JSON
  pm connectors inspect freightview --json

AGENT WORKFLOW
  - Run pm connectors inspect freightview before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
