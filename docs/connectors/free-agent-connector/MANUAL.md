# pm connectors inspect free-agent-connector

```text
NAME
  pm connectors inspect free-agent-connector - FreeAgent connector manual

SYNOPSIS
  pm connectors inspect free-agent-connector
  pm connectors inspect free-agent-connector --json
  pm credentials add <name> --connector free-agent-connector [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads FreeAgent contacts, invoices, bills, projects, and tasks through the FreeAgent v2 REST API using OAuth2 refresh-token authentication.

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
  pm connectors inspect free-agent-connector

  # Inspect as structured JSON
  pm connectors inspect free-agent-connector --json

AGENT WORKFLOW
  - Run pm connectors inspect free-agent-connector before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
