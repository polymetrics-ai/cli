# pm connectors inspect chartmogul

```text
NAME
  pm connectors inspect chartmogul - ChartMogul connector manual

SYNOPSIS
  pm connectors inspect chartmogul
  pm connectors inspect chartmogul --json
  pm credentials add <name> --connector chartmogul [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes ChartMogul customers, contacts, subscription activities, plans, invoices, tasks, customer-count metrics, and account details through the ChartMogul REST API.

ICON
  asset: icons/chartmogul.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://dev.chartmogul.com/reference

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  api_key (secret)

ETL STREAMS
  customers:
    primary key: uuid
    cursor: customer-since
    fields: arr(), billing-system-type(), city(), company(), country(), currency(), customer-since(), email(), external_id(), mrr(), name(), status(), uuid()
  activities:
    primary key: uuid
    cursor: date
    fields: activity-arr(), activity-mrr(), activity-mrr-movement(), currency(), customer-external-id(), customer-name(), customer-uuid(), date(), description(), plan-external-id(), subscription-external-id(), type(), uuid()
  customer_count:
    primary key: date
    cursor: date
    fields: customers(), date(), percentage-change()
  account:
    primary key: uuid
    fields: currency(), name(), time_zone(), uuid(), week_start_on()
  plans:
    primary key: uuid
    fields: data_source_uuid(), external_id(), interval_count(), interval_unit(), name(), uuid()
  contacts:
    primary key: uuid
    fields: customer_external_id(), customer_uuid(), data_source_uuid(), email(), external_id(), first_name(), last_name(), last_seen(), phone(), title(), uuid()
  tasks:
    primary key: task_uuid
    cursor: updated_at
    fields: assignee(), completed_at(), created_at(), customer_uuid(), due_date(), task_details(), task_uuid(), updated_at()
  invoices:
    primary key: uuid
    fields: currency(), customer_uuid(), date(), due_date(), external_id(), uuid()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    risk: external mutation; approval required
  update_customer:
    endpoint: PUT /customers/{{ record.uuid }}
    required fields: uuid
    risk: external mutation; approval required

SECURITY
  read risk: external ChartMogul API read of customer, contact, CRM-task, plan, invoice, and subscription-metrics data
  write risk: external mutation of ChartMogul customer records; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect chartmogul

  # Inspect as structured JSON
  pm connectors inspect chartmogul --json

AGENT WORKFLOW
  - Run pm connectors inspect chartmogul before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
