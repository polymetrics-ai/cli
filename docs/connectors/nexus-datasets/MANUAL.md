# pm connectors inspect nexus-datasets

```text
NAME
  pm connectors inspect nexus-datasets - Infor Nexus Datasets connector manual

SYNOPSIS
  pm connectors inspect nexus-datasets
  pm connectors inspect nexus-datasets --json
  pm credentials add <name> --connector nexus-datasets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads records from a configured Infor Nexus export dataset through the Infor Nexus Data API (v3.1) using HMAC authentication. Read-only.

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
  pm connectors inspect nexus-datasets

  # Inspect as structured JSON
  pm connectors inspect nexus-datasets --json

AGENT WORKFLOW
  - Run pm connectors inspect nexus-datasets before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
