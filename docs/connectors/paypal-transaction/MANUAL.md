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

ICON
  asset: icons/paypal.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.paypal.com/api/rest/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  end_date
  max_pages
  mode
  start_date
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  transactions:
    primary key: transaction_id
    cursor: transaction_initiation_date
    fields: amount(), currency_code(), fee_amount(), paypal_account_id(), transaction_event_code(), transaction_id(), transaction_initiation_date(), transaction_status(), transaction_updated_date()
  balances:
    primary key: currency
    fields: available_value(), currency(), primary(), total_currency_code(), total_value(), withheld_value()
  products:
    primary key: id
    fields: category(), create_time(), description(), id(), name(), type()
  disputes:
    primary key: dispute_id
    cursor: update_time
    fields: create_time(), dispute_amount_currency_code(), dispute_amount_value(), dispute_id(), dispute_state(), reason(), status(), update_time()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external PayPal REST API read of transaction, balance, catalog, and dispute data
  approval: none; read-only, no reverse-ETL write surface
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
