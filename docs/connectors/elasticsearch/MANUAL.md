# pm connectors inspect elasticsearch

```text
NAME
  pm connectors inspect elasticsearch - Elasticsearch connector manual

SYNOPSIS
  pm connectors inspect elasticsearch
  pm connectors inspect elasticsearch --json
  pm credentials add <name> --connector elasticsearch [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Elasticsearch index metadata and documents through the REST API.

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
  pm connectors inspect elasticsearch

  # Inspect as structured JSON
  pm connectors inspect elasticsearch --json

AGENT WORKFLOW
  - Run pm connectors inspect elasticsearch before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
