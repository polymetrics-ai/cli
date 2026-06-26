# pm connectors inspect employment-hero

```text
NAME
  pm connectors inspect employment-hero - Employment Hero connector manual

SYNOPSIS
  pm connectors inspect employment-hero
  pm connectors inspect employment-hero --json
  pm credentials add <name> --connector employment-hero [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Employment Hero organisations, employees, leave requests, and teams through the Employment Hero REST API. Read-only (the API is full-refresh only).

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
  pm connectors inspect employment-hero

  # Inspect as structured JSON
  pm connectors inspect employment-hero --json

AGENT WORKFLOW
  - Run pm connectors inspect employment-hero before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
