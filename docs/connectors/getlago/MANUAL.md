# pm connectors inspect getlago

```text
NAME
  pm connectors inspect getlago - Lago connector manual

SYNOPSIS
  pm connectors inspect getlago
  pm connectors inspect getlago --json
  pm credentials add <name> --connector getlago [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Lago customers, invoices, subscriptions, plans, and billable metrics through the Lago REST API.

ICON
  id: getlago
  asset: icons/getlago.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://doc.getlago.com/api-reference/intro

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_url
  max_pages
  mode
  page_size
  api_key (secret) (required)

ETL STREAMS
  customers:
    primary key: lago_id
    cursor: created_at
    fields: country(string), created_at(string), currency(string), customer_type(string), email(string), external_id(string), lago_id(string), name(string), sequential_id(integer), slug(string), updated_at(string)
  invoices:
    primary key: lago_id
    cursor: created_at
    fields: created_at(string), currency(string), fees_amount_cents(integer), invoice_type(string), issuing_date(string), lago_id(string), number(string), payment_status(string), status(string), taxes_amount_cents(integer), total_amount_cents(integer), updated_at(string)
  subscriptions:
    primary key: lago_id
    cursor: created_at
    fields: billing_time(string), created_at(string), external_customer_id(string), external_id(string), lago_customer_id(string), lago_id(string), plan_code(string), started_at(string), status(string), terminated_at(string)
  plans:
    primary key: lago_id
    cursor: created_at
    fields: amount_cents(integer), amount_currency(string), code(string), created_at(string), interval(string), lago_id(string), name(string), pay_in_advance(boolean), trial_period(number)
  billable_metrics:
    primary key: lago_id
    cursor: created_at
    fields: aggregation_type(string), code(string), created_at(string), field_name(string), lago_id(string), name(string), recurring(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Lago API read of billing and subscription data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect getlago

  # Inspect as structured JSON
  pm connectors inspect getlago --json

AGENT WORKFLOW
  - Run pm connectors inspect getlago before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
