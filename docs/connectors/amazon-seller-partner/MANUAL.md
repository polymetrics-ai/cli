# pm connectors inspect amazon-seller-partner

```text
NAME
  pm connectors inspect amazon-seller-partner - Amazon Seller Partner connector manual

SYNOPSIS
  pm connectors inspect amazon-seller-partner
  pm connectors inspect amazon-seller-partner --json
  pm credentials add <name> --connector amazon-seller-partner [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Amazon Selling Partner API orders, order items, FBA inventory summaries, and financial event groups via Login with Amazon (LWA) authentication.

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
  pm connectors inspect amazon-seller-partner

  # Inspect as structured JSON
  pm connectors inspect amazon-seller-partner --json

AGENT WORKFLOW
  - Run pm connectors inspect amazon-seller-partner before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
