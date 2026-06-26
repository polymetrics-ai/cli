# pm connectors inspect hibob

```text
NAME
  pm connectors inspect hibob - HiBob connector manual

SYNOPSIS
  pm connectors inspect hibob
  pm connectors inspect hibob --json
  pm credentials add <name> --connector hibob [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads HiBob HR data: employee profiles, company named lists, and people field definitions via the HiBob REST API (read-only).

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
  pm connectors inspect hibob

  # Inspect as structured JSON
  pm connectors inspect hibob --json

AGENT WORKFLOW
  - Run pm connectors inspect hibob before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
