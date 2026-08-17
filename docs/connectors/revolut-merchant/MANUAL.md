# pm connectors inspect revolut-merchant

```text
NAME
  pm connectors inspect revolut-merchant - Revolut Merchant connector manual

SYNOPSIS
  pm connectors inspect revolut-merchant
  pm connectors inspect revolut-merchant --json
  pm credentials add <name> --connector revolut-merchant [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Revolut Merchant orders, customers, settlements, and payment links through the REST API.

ICON
  asset: icons/revolut.svg
  source: official
  review_status: official_verified
  review_url: https://developer.revolut.com/docs/guides/merchant/reference/api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  customer_id
  from_created_date
  state
  to_created_date
  api_key (secret)

ETL STREAMS
  orders:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), id(), state(), stream()
  customers:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), full_name(), id(), stream()
  settlements:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), id(), stream()
  payment_links:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), id(), state(), stream()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Revolut Merchant API read of order, customer, settlement, and payment-link data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect revolut-merchant

  # Inspect as structured JSON
  pm connectors inspect revolut-merchant --json

AGENT WORKFLOW
  - Run pm connectors inspect revolut-merchant before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
