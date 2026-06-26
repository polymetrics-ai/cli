# pm connectors inspect tinyemail

```text
NAME
  pm connectors inspect tinyemail - TinyEmail connector manual

SYNOPSIS
  pm connectors inspect tinyemail
  pm connectors inspect tinyemail --json
  pm credentials add <name> --connector tinyemail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads subscribers, lists, and campaigns from the tinyEmail API.

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
  pm connectors inspect tinyemail

  # Inspect as structured JSON
  pm connectors inspect tinyemail --json

AGENT WORKFLOW
  - Run pm connectors inspect tinyemail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
