# pm connectors inspect paddle

```text
NAME
  pm connectors inspect paddle - Paddle connector manual

SYNOPSIS
  pm connectors inspect paddle
  pm connectors inspect paddle --json
  pm credentials add <name> --connector paddle [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Paddle customers, subscriptions, transactions, and products through the Paddle REST API.

ICON
  asset: icons/paddle.svg
  source: official
  review_status: official_verified
  review_url: https://developer.paddle.com/api-reference/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  transactions:
    primary key: id
    cursor: created_at
    fields: created_at(), currency_code(), customer_id(), id(), status(), subscription_id()
  customers:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), id(), name()
  subscriptions:
    primary key: id
    cursor: created_at
    fields: created_at(), customer_id(), id(), status()
  products:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Paddle API read of customer, subscription, transaction, and product data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect paddle

  # Inspect as structured JSON
  pm connectors inspect paddle --json

AGENT WORKFLOW
  - Run pm connectors inspect paddle before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
