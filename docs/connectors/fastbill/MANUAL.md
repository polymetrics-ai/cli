# pm connectors inspect fastbill

```text
NAME
  pm connectors inspect fastbill - FastBill connector manual

SYNOPSIS
  pm connectors inspect fastbill
  pm connectors inspect fastbill --json
  pm credentials add <name> --connector fastbill [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads FastBill customers, invoices, products, recurring invoices, and revenues through the FastBill JSON API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  id: fastbill
  asset: icons/fastbill.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://apidocs.fastbill.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  username (required)
  api_key (secret) (required)

ETL STREAMS
  customers:
    primary key: customer_id
    fields: country_code(string), created(string), currency_code(string), customer_id(string), customer_number(string), customer_type(string), email(string), first_name(string), last_name(string), organization(string), phone(string)
  invoices:
    primary key: invoice_id
    fields: currency_code(string), customer_id(string), due_date(string), invoice_date(string), invoice_id(string), invoice_number(string), is_canceled(string), sub_total(string), total(string), type(string), vat_total(string)
  products:
    primary key: article_number
    fields: article_number(string), currency_code(string), description(string), is_greedy(string), title(string), unit_price(string), vat_percent(string)
  recurring_invoices:
    primary key: invoice_id
    fields: currency_code(string), customer_id(string), due_date(string), invoice_date(string), invoice_id(string), invoice_number(string), is_canceled(string), sub_total(string), total(string), type(string), vat_total(string)
  revenues:
    primary key: invoice_id
    fields: currency_code(string), customer_id(string), invoice_date(string), invoice_id(string), invoice_number(string), total(string), vat_total(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external FastBill API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect fastbill

  # Inspect as structured JSON
  pm connectors inspect fastbill --json

AGENT WORKFLOW
  - Run pm connectors inspect fastbill before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
