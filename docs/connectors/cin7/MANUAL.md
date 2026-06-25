# pm connectors inspect cin7

```text
NAME
  pm connectors inspect cin7 - Cin7 connector manual

SYNOPSIS
  pm connectors inspect cin7
  pm connectors inspect cin7 --json
  pm credentials add <name> --connector cin7 [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Cin7 Core (DEAR Inventory) products, customers, suppliers, sales, and purchases through the Cin7 Core REST API.

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
  pm connectors inspect cin7

  # Inspect as structured JSON
  pm connectors inspect cin7 --json

AGENT WORKFLOW
  - Run pm connectors inspect cin7 before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
