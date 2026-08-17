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
  api_key (secret)

ETL STREAMS
  customers:
    primary key: lago_id
    cursor: created_at
    fields: country(), created_at(), currency(), customer_type(), email(), external_id(), lago_id(), name(), sequential_id(), slug(), updated_at()
  invoices:
    primary key: lago_id
    cursor: created_at
    fields: created_at(), currency(), fees_amount_cents(), invoice_type(), issuing_date(), lago_id(), number(), payment_status(), status(), taxes_amount_cents(), total_amount_cents(), updated_at()
  subscriptions:
    primary key: lago_id
    cursor: created_at
    fields: billing_time(), created_at(), external_customer_id(), external_id(), lago_customer_id(), lago_id(), plan_code(), started_at(), status(), terminated_at()
  plans:
    primary key: lago_id
    cursor: created_at
    fields: amount_cents(), amount_currency(), code(), created_at(), interval(), lago_id(), name(), pay_in_advance(), trial_period()
  billable_metrics:
    primary key: lago_id
    cursor: created_at
    fields: aggregation_type(), code(), created_at(), field_name(), lago_id(), name(), recurring()

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
