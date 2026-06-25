# pm connectors inspect appfigures

```text
NAME
  pm connectors inspect appfigures - Appfigures connector manual

SYNOPSIS
  pm connectors inspect appfigures
  pm connectors inspect appfigures --json
  pm credentials add <name> --connector appfigures [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Appfigures app-store analytics: reviews, products, sales and ratings reports, and store categories via the Appfigures v2 REST API.

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
  pm connectors inspect appfigures

  # Inspect as structured JSON
  pm connectors inspect appfigures --json

AGENT WORKFLOW
  - Run pm connectors inspect appfigures before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
