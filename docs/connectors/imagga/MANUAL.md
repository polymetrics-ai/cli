# pm connectors inspect imagga

```text
NAME
  pm connectors inspect imagga - Imagga connector manual

SYNOPSIS
  pm connectors inspect imagga
  pm connectors inspect imagga --json
  pm credentials add <name> --connector imagga [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Imagga image-recognition results (tags, categories, colors, face detections) and account usage via the Imagga REST API. Read-only.

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
  pm connectors inspect imagga

  # Inspect as structured JSON
  pm connectors inspect imagga --json

AGENT WORKFLOW
  - Run pm connectors inspect imagga before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
