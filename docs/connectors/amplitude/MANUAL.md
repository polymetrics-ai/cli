# pm connectors inspect amplitude

```text
NAME
  pm connectors inspect amplitude - Amplitude connector manual

SYNOPSIS
  pm connectors inspect amplitude
  pm connectors inspect amplitude --json
  pm credentials add <name> --connector amplitude [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Amplitude behavioral cohorts, chart annotations, and event types through the Amplitude Analytics REST API.

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
  pm connectors inspect amplitude

  # Inspect as structured JSON
  pm connectors inspect amplitude --json

AGENT WORKFLOW
  - Run pm connectors inspect amplitude before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
