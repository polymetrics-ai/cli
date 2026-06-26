# pm connectors inspect visma-economic

```text
NAME
  pm connectors inspect visma-economic - Visma e-conomic connector manual

SYNOPSIS
  pm connectors inspect visma-economic
  pm connectors inspect visma-economic --json
  pm credentials add <name> --connector visma-economic [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Visma e-conomic customers.

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
  pm connectors inspect visma-economic

  # Inspect as structured JSON
  pm connectors inspect visma-economic --json

AGENT WORKFLOW
  - Run pm connectors inspect visma-economic before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
