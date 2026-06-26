# pm connectors inspect flowlu

```text
NAME
  pm connectors inspect flowlu - Flowlu connector manual

SYNOPSIS
  pm connectors inspect flowlu
  pm connectors inspect flowlu --json
  pm credentials add <name> --connector flowlu [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Flowlu CRM accounts, leads, tasks, projects, invoices, and agile issues through the Flowlu REST API.

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
  pm connectors inspect flowlu

  # Inspect as structured JSON
  pm connectors inspect flowlu --json

AGENT WORKFLOW
  - Run pm connectors inspect flowlu before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
