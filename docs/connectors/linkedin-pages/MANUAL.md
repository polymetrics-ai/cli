# pm connectors inspect linkedin-pages

```text
NAME
  pm connectors inspect linkedin-pages - LinkedIn Pages connector manual

SYNOPSIS
  pm connectors inspect linkedin-pages
  pm connectors inspect linkedin-pages --json
  pm credentials add <name> --connector linkedin-pages [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads LinkedIn organization (company page) profile, follower statistics, share statistics, and total follower count through the LinkedIn Community Management REST API.

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
  pm connectors inspect linkedin-pages

  # Inspect as structured JSON
  pm connectors inspect linkedin-pages --json

AGENT WORKFLOW
  - Run pm connectors inspect linkedin-pages before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
