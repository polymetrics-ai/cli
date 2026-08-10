# pm connectors inspect zoho-invoice

```text
NAME
  pm connectors inspect zoho-invoice - Zoho Invoice connector manual

SYNOPSIS
  pm connectors inspect zoho-invoice
  pm connectors inspect zoho-invoice --json
  pm credentials add <name> --connector zoho-invoice [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoho Invoice customers, invoices, and payments through the Zoho Invoice REST API.

ICON
  id: simple-icons-zoho-invoice
  asset: icons/simple-icons/zoho-invoice.svg
  title: Zoho
  simple_icon_slug: zoho
  simple_icon_hex: E42527
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Zoho
  match: curated-alias
  matched_by: zoho

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  organization_id
  page_size
  access_token (secret) (required)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: company_name(string), created_time(string), currency_code(string), customer_id(string), customer_name(string), customer_type(string), email(string), id(string), last_modified_time(string), outstanding_receivable_amount(number), phone(string), status(string), updated_at(string)
  invoices:
    primary key: id
    cursor: updated_at
    fields: balance(number), created_time(string), currency_code(string), customer_id(string), customer_name(string), date(string), due_date(string), id(string), invoice_id(string), invoice_number(string), last_modified_time(string), status(string), total(number), updated_at(string)
  payments:
    primary key: id
    cursor: updated_at
    fields: amount(number), created_time(string), currency_code(string), customer_id(string), customer_name(string), date(string), id(string), invoice_numbers(string), last_modified_time(string), payment_id(string), payment_mode(string), payment_number(string), reference_number(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Zoho Invoice API read of customer/invoice/payment data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoho-invoice

  # Inspect as structured JSON
  pm connectors inspect zoho-invoice --json

AGENT WORKFLOW
  - Run pm connectors inspect zoho-invoice before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
