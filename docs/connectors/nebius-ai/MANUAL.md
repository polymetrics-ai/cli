# pm connectors inspect nebius-ai

```text
NAME
  pm connectors inspect nebius-ai - Nebius AI connector manual

SYNOPSIS
  pm connectors inspect nebius-ai
  pm connectors inspect nebius-ai --json
  pm credentials add <name> --connector nebius-ai [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Nebius AI Studio models, files, and batch jobs through the OpenAI-compatible Nebius AI Studio REST API.

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
  pm connectors inspect nebius-ai

  # Inspect as structured JSON
  pm connectors inspect nebius-ai --json

AGENT WORKFLOW
  - Run pm connectors inspect nebius-ai before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
