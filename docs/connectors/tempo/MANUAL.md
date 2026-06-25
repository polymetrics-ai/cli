# pm connectors inspect tempo

```text
NAME
  pm connectors inspect tempo - Tempo connector manual

SYNOPSIS
  pm connectors inspect tempo
  pm connectors inspect tempo --json
  pm credentials add <name> --connector tempo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Tempo accounts, customers, worklogs, and workload schemes through the Tempo Cloud REST API v4.

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
  pm connectors inspect tempo

  # Inspect as structured JSON
  pm connectors inspect tempo --json

AGENT WORKFLOW
  - Run pm connectors inspect tempo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
