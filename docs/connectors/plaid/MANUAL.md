# pm connectors inspect plaid

```text
NAME
  pm connectors inspect plaid - Plaid connector manual

SYNOPSIS
  pm connectors inspect plaid
  pm connectors inspect plaid --json
  pm credentials add <name> --connector plaid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Plaid institutions and category metadata through read-only POST endpoints.

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
  pm connectors inspect plaid

  # Inspect as structured JSON
  pm connectors inspect plaid --json

AGENT WORKFLOW
  - Run pm connectors inspect plaid before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
