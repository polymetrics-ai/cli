# pm connectors inspect awin-advertiser

```text
NAME
  pm connectors inspect awin-advertiser - Awin Advertiser connector manual

SYNOPSIS
  pm connectors inspect awin-advertiser
  pm connectors inspect awin-advertiser --json
  pm credentials add <name> --connector awin-advertiser [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Awin advertiser transactions, publisher-aggregated performance reports, and publisher relationships through the Awin Advertiser REST API.

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
  pm connectors inspect awin-advertiser

  # Inspect as structured JSON
  pm connectors inspect awin-advertiser --json

AGENT WORKFLOW
  - Run pm connectors inspect awin-advertiser before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
