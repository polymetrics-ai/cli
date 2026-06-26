# pm connectors inspect mux

```text
NAME
  pm connectors inspect mux - Mux connector manual

SYNOPSIS
  pm connectors inspect mux
  pm connectors inspect mux --json
  pm credentials add <name> --connector mux [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mux Video assets, live streams, direct uploads, and system signing keys through the Mux REST API using HTTP Basic authentication.

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
  pm connectors inspect mux

  # Inspect as structured JSON
  pm connectors inspect mux --json

AGENT WORKFLOW
  - Run pm connectors inspect mux before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
