# pm connectors inspect pinterest

```text
NAME
  pm connectors inspect pinterest - Pinterest connector manual

SYNOPSIS
  pm connectors inspect pinterest
  pm connectors inspect pinterest --json
  pm credentials add <name> --connector pinterest [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pinterest ad accounts, boards, campaigns, ad groups, and audiences through the Pinterest API v5 (OAuth2 refresh-token auth). Read-only.

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
  pm connectors inspect pinterest

  # Inspect as structured JSON
  pm connectors inspect pinterest --json

AGENT WORKFLOW
  - Run pm connectors inspect pinterest before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
