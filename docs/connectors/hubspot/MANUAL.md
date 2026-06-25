# pm connectors inspect hubspot

```text
NAME
  pm connectors inspect hubspot - HubSpot connector manual

SYNOPSIS
  pm connectors inspect hubspot
  pm connectors inspect hubspot --json
  pm credentials add <name> --connector hubspot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads HubSpot CRM contacts, companies, deals, and tickets, and writes approved reverse ETL contact actions through the HubSpot CRM v3 REST API.

CAPABILITIES
  check=true catalog=true read=true write=true query=false
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
  pm connectors inspect hubspot

  # Inspect as structured JSON
  pm connectors inspect hubspot --json

AGENT WORKFLOW
  - Run pm connectors inspect hubspot before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
