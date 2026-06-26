# pm connectors inspect kissmetrics

```text
NAME
  pm connectors inspect kissmetrics - Kissmetrics connector manual

SYNOPSIS
  pm connectors inspect kissmetrics
  pm connectors inspect kissmetrics --json
  pm credentials add <name> --connector kissmetrics [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Kissmetrics products, reports, events, and properties through the Kissmetrics query API using HTTP Basic authentication.

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
  pm connectors inspect kissmetrics

  # Inspect as structured JSON
  pm connectors inspect kissmetrics --json

AGENT WORKFLOW
  - Run pm connectors inspect kissmetrics before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
