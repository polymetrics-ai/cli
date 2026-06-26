# pm connectors inspect papersign

```text
NAME
  pm connectors inspect papersign - PaperSign connector manual

SYNOPSIS
  pm connectors inspect papersign
  pm connectors inspect papersign --json
  pm credentials add <name> --connector papersign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PaperSign documents, templates, and recipients through the REST API.

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
  pm connectors inspect papersign

  # Inspect as structured JSON
  pm connectors inspect papersign --json

AGENT WORKFLOW
  - Run pm connectors inspect papersign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
