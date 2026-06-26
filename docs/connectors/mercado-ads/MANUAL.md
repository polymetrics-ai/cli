# pm connectors inspect mercado-ads

```text
NAME
  pm connectors inspect mercado-ads - Mercado Ads connector manual

SYNOPSIS
  pm connectors inspect mercado-ads
  pm connectors inspect mercado-ads --json
  pm credentials add <name> --connector mercado-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mercado Ads brand, display, and product advertisers and daily campaign metrics from the Mercado Libre Advertising API.

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
  pm connectors inspect mercado-ads

  # Inspect as structured JSON
  pm connectors inspect mercado-ads --json

AGENT WORKFLOW
  - Run pm connectors inspect mercado-ads before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
