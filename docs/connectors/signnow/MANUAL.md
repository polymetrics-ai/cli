# pm connectors inspect signnow

```text
NAME
  pm connectors inspect signnow - signNow connector manual

SYNOPSIS
  pm connectors inspect signnow
  pm connectors inspect signnow --json
  pm credentials add <name> --connector signnow [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads signNow documents, templates, and users through the signNow REST API.

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
  pm connectors inspect signnow

  # Inspect as structured JSON
  pm connectors inspect signnow --json

AGENT WORKFLOW
  - Run pm connectors inspect signnow before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
