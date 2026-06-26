# pm connectors inspect mailtrap

```text
NAME
  pm connectors inspect mailtrap - Mailtrap connector manual

SYNOPSIS
  pm connectors inspect mailtrap
  pm connectors inspect mailtrap --json
  pm credentials add <name> --connector mailtrap [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailtrap accounts, inboxes, projects, and sending domains through the Mailtrap account-management REST API.

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
  pm connectors inspect mailtrap

  # Inspect as structured JSON
  pm connectors inspect mailtrap --json

AGENT WORKFLOW
  - Run pm connectors inspect mailtrap before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
