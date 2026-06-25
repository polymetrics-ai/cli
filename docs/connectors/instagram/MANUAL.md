# pm connectors inspect instagram

```text
NAME
  pm connectors inspect instagram - Instagram connector manual

SYNOPSIS
  pm connectors inspect instagram
  pm connectors inspect instagram --json
  pm credentials add <name> --connector instagram [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Instagram Business/Creator account profile, media, stories, and account insights through the Facebook Graph API.

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
  pm connectors inspect instagram

  # Inspect as structured JSON
  pm connectors inspect instagram --json

AGENT WORKFLOW
  - Run pm connectors inspect instagram before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
