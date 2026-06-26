# pm connectors inspect rki-covid

```text
NAME
  pm connectors inspect rki-covid - RKI COVID connector manual

SYNOPSIS
  pm connectors inspect rki-covid
  pm connectors inspect rki-covid --json
  pm credentials add <name> --connector rki-covid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public Germany COVID case, state, district, and history data derived from RKI reports.

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
  pm connectors inspect rki-covid

  # Inspect as structured JSON
  pm connectors inspect rki-covid --json

AGENT WORKFLOW
  - Run pm connectors inspect rki-covid before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
