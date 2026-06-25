# pm connectors inspect paypal-transaction

```text
NAME
  pm connectors inspect paypal-transaction - PayPal Transaction connector manual

SYNOPSIS
  pm connectors inspect paypal-transaction
  pm connectors inspect paypal-transaction --json
  pm credentials add <name> --connector paypal-transaction [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PayPal transactions, balances, catalog products, and customer disputes through the PayPal REST API using OAuth 2.0 client-credentials auth.

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
  pm connectors inspect paypal-transaction

  # Inspect as structured JSON
  pm connectors inspect paypal-transaction --json

AGENT WORKFLOW
  - Run pm connectors inspect paypal-transaction before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
