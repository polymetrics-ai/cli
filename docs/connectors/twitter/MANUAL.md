# pm connectors inspect twitter

```text
NAME
  pm connectors inspect twitter - Twitter connector manual

SYNOPSIS
  pm connectors inspect twitter
  pm connectors inspect twitter --json
  pm credentials add <name> --connector twitter [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads tweets and their authors matching a search query from the Twitter (X) API v2 recent search endpoint using an App-only Bearer token.

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
  pm connectors inspect twitter

  # Inspect as structured JSON
  pm connectors inspect twitter --json

AGENT WORKFLOW
  - Run pm connectors inspect twitter before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
