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
  id: paypal
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
  start_date (required)
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  transactions:
    primary key: transaction_id
    cursor: transaction_initiation_date
    fields: amount(string), currency_code(string), fee_amount(string), paypal_account_id(string), transaction_event_code(string), transaction_id(string), transaction_initiation_date(string), transaction_status(string), transaction_updated_date(string)
  balances:
    primary key: currency
    fields: available_value(string), currency(string), primary(boolean), total_currency_code(string), total_value(string), withheld_value(string)
  products:
    primary key: id
    fields: category(string), create_time(string), description(string), id(string), name(string), type(string)
  disputes:
    primary key: dispute_id
    cursor: update_time
    fields: create_time(string), dispute_amount_currency_code(string), dispute_amount_value(string), dispute_id(string), dispute_state(string), reason(string), status(string), update_time(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external PayPal REST API read of transaction, balance, catalog, and dispute data
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read declared PayPal Transaction streams and bounded PayPal REST resources.
  Usage: pm paypal-transaction <command> [flags]
  Source CLI: PayPal REST API (https://github.com/paypal/paypal-rest-api-specifications)
  Global flags:
    --credential (string): Named PayPal credential; secrets are loaded from the credential store.
    --json (boolean): Emit machine-readable JSON output.
    --max-bytes (integer): Clamp direct-read response size; these operations are capped at 1 MiB.
  PayPal Invoicing direct reads
  PayPal webhook direct reads
  PayPal payment-experience direct reads
  Other Commands
    invoicing connections list - List PayPal invoicing accounting-sync merchant connections. [intent=direct_read availability=implemented operation=paypal-transaction.invoicing.connections.get]; approval: none; risk: bounded read; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor
    webhooks lookup list - List PayPal webhook lookups. [intent=direct_read availability=implemented operation=paypal-transaction.webhooks.lookup.list]; approval: none; risk: bounded read; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor
    webhooks event-types list - List PayPal webhook event types. [intent=direct_read availability=implemented operation=paypal-transaction.webhooks.event-types.list]; approval: none; risk: bounded read; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor
    web-profiles list - List PayPal payment-experience web profiles. [intent=direct_read availability=implemented operation=paypal-transaction.web-profiles.list]; approval: none; risk: bounded read; the response is capped at 1 MiB and redacted before JSON output; flags: --page, --page-cursor

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
