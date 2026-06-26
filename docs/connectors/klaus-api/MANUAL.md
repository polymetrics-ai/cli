# pm connectors inspect klaus-api

```text
NAME
  pm connectors inspect klaus-api - Klaus API connector manual

SYNOPSIS
  pm connectors inspect klaus-api
  pm connectors inspect klaus-api --json
  pm credentials add <name> --connector klaus-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Klaus (Zendesk QA) users, rating categories, and review conversations through the Klaus public REST API.

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
  pm connectors inspect klaus-api

  # Inspect as structured JSON
  pm connectors inspect klaus-api --json

AGENT WORKFLOW
  - Run pm connectors inspect klaus-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
