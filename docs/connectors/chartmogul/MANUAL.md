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
  id: chartmogul
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
  api_key (secret) (required)

ETL STREAMS
  customers:
    primary key: uuid
    cursor: customer-since
    fields: arr(integer), billing-system-type(string), city(string), company(string), country(string), currency(string), customer-since(string), email(string), external_id(string), mrr(integer), name(string), status(string), uuid(string)
  activities:
    primary key: uuid
    cursor: date
    fields: activity-arr(integer), activity-mrr(integer), activity-mrr-movement(integer), currency(string), customer-external-id(string), customer-name(string), customer-uuid(string), date(string), description(string), plan-external-id(string), subscription-external-id(string), type(string), uuid(string)
  customer_count:
    primary key: date
    cursor: date
    fields: customers(integer), date(string), percentage-change(number)
  account:
    primary key: uuid
    fields: currency(string), name(string), time_zone(string), uuid(string), week_start_on(string)
  plans:
    primary key: uuid
    fields: data_source_uuid(string), external_id(string), interval_count(integer), interval_unit(string), name(string), uuid(string)
  contacts:
    primary key: uuid
    fields: customer_external_id(string), customer_uuid(string), data_source_uuid(string), email(string), external_id(string), first_name(string), last_name(string), last_seen(string), phone(string), title(string), uuid(string)
  tasks:
    primary key: task_uuid
    cursor: updated_at
    fields: assignee(string), completed_at(string), created_at(string), customer_uuid(string), due_date(string), task_details(string), task_uuid(string), updated_at(string)
  invoices:
    primary key: uuid
    fields: currency(string), customer_uuid(string), date(string), due_date(string), external_id(string), uuid(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    required fields: data_source_uuid, external_id
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
