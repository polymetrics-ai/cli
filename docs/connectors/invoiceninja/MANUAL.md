# pm connectors inspect invoiceninja

```text
NAME
  pm connectors inspect invoiceninja - Invoice Ninja connector manual

SYNOPSIS
  pm connectors inspect invoiceninja
  pm connectors inspect invoiceninja --json
  pm credentials add <name> --connector invoiceninja [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Invoice Ninja clients, invoices, products, payments, and quotes through the Invoice Ninja v5 REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  api_key (secret)

ETL STREAMS
  clients:
    primary key: id
    fields: archived_at(), balance(), created_at(), currency_id(), display_name(), id(), is_deleted(), name(), number(), paid_to_date(), phone(), updated_at(), vat_number(), website()
  invoices:
    primary key: id
    fields: amount(), balance(), client_id(), created_at(), currency_id(), date(), due_date(), id(), is_deleted(), number(), paid_to_date(), status_id(), updated_at()
  products:
    primary key: id
    fields: cost(), created_at(), id(), is_deleted(), notes(), price(), product_key(), quantity(), tax_name1(), tax_rate1(), updated_at()
  payments:
    primary key: id
    fields: amount(), applied(), client_id(), created_at(), currency_id(), date(), id(), is_deleted(), number(), refunded(), status_id(), transaction_reference(), updated_at()
  quotes:
    primary key: id
    fields: amount(), balance(), client_id(), created_at(), currency_id(), date(), due_date(), id(), is_deleted(), number(), status_id(), updated_at(), valid_until()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Invoice Ninja API read of client and billing data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect invoiceninja

  # Inspect as structured JSON
  pm connectors inspect invoiceninja --json

AGENT WORKFLOW
  - Run pm connectors inspect invoiceninja before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
