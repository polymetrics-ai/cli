# pm connectors inspect uppromote

```text
NAME
  pm connectors inspect uppromote - UpPromote connector manual

SYNOPSIS
  pm connectors inspect uppromote
  pm connectors inspect uppromote --json
  pm credentials add <name> --connector uppromote [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads affiliates from the UpPromote API.

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
  pm connectors inspect uppromote

  # Inspect as structured JSON
  pm connectors inspect uppromote --json

AGENT WORKFLOW
  - Run pm connectors inspect uppromote before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
