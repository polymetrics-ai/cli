# pm connectors inspect flexport

```text
NAME
  pm connectors inspect flexport - Flexport connector manual

SYNOPSIS
  pm connectors inspect flexport
  pm connectors inspect flexport --json
  pm credentials add <name> --connector flexport [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Flexport companies, locations, products, invoices, and shipments through the Flexport REST API. Read-only.

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
  pm connectors inspect flexport

  # Inspect as structured JSON
  pm connectors inspect flexport --json

AGENT WORKFLOW
  - Run pm connectors inspect flexport before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
