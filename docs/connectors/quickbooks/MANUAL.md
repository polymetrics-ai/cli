# pm connectors inspect quickbooks

```text
NAME
  pm connectors inspect quickbooks - QuickBooks connector manual

SYNOPSIS
  pm connectors inspect quickbooks
  pm connectors inspect quickbooks --json
  pm credentials add <name> --connector quickbooks [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads QuickBooks Online customers, invoices, payments, accounts, and vendors through the v3 Query API via the OAuth 2.0 refresh-token grant. Read-only.

ICON
  asset: icons/quickbooks.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  page_size
  sandbox
  start_date
  token_url
  client_id (secret)
  client_secret (secret)
  realm_id (secret)
  refresh_token (secret)

ETL STREAMS
  customers:
    primary key: id
    fields: active(), balance(), display_name(), id()
  invoices:
    primary key: id
    fields: balance(), customer_ref(), doc_number(), id(), total_amt()
  payments:
    primary key: id
    fields: customer_ref(), id(), total_amt(), txn_date()
  accounts:
    primary key: id
    fields: account_type(), classification(), id(), name()
  vendors:
    primary key: id
    fields: active(), balance(), display_name(), id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external QuickBooks Online v3 Query API read of customer/invoice/payment/account/vendor entities
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect quickbooks

  # Inspect as structured JSON
  pm connectors inspect quickbooks --json

AGENT WORKFLOW
  - Run pm connectors inspect quickbooks before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
