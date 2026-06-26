# pm connectors inspect scryfall

```text
NAME
  pm connectors inspect scryfall - Scryfall connector manual

SYNOPSIS
  pm connectors inspect scryfall
  pm connectors inspect scryfall --json
  pm credentials add <name> --connector scryfall [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads cards and sets from the public Scryfall API. Read-only and credential-free.

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
  pm connectors inspect scryfall

  # Inspect as structured JSON
  pm connectors inspect scryfall --json

AGENT WORKFLOW
  - Run pm connectors inspect scryfall before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
