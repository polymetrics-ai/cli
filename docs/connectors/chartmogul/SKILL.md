---
name: pm-chartmogul
description: ChartMogul connector knowledge and safe action guide.
---

# pm-chartmogul

## Purpose

Reads and writes ChartMogul customers, contacts, subscription activities, plans, invoices, tasks, customer-count metrics, and account details through the ChartMogul REST API.

## Icon

- asset: icons/chartmogul.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://dev.chartmogul.com/reference

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- api_key (secret)

## ETL Streams

- customers:
  - primary key: uuid
  - cursor: customer-since
  - fields: arr(), billing-system-type(), city(), company(), country(), currency(), customer-since(), email(), external_id(), mrr(), name(), status(), uuid()
- activities:
  - primary key: uuid
  - cursor: date
  - fields: activity-arr(), activity-mrr(), activity-mrr-movement(), currency(), customer-external-id(), customer-name(), customer-uuid(), date(), description(), plan-external-id(), subscription-external-id(), type(), uuid()
- customer_count:
  - primary key: date
  - cursor: date
  - fields: customers(), date(), percentage-change()
- account:
  - primary key: uuid
  - fields: currency(), name(), time_zone(), uuid(), week_start_on()
- plans:
  - primary key: uuid
  - fields: data_source_uuid(), external_id(), interval_count(), interval_unit(), name(), uuid()
- contacts:
  - primary key: uuid
  - fields: customer_external_id(), customer_uuid(), data_source_uuid(), email(), external_id(), first_name(), last_name(), last_seen(), phone(), title(), uuid()
- tasks:
  - primary key: task_uuid
  - cursor: updated_at
  - fields: assignee(), completed_at(), created_at(), customer_uuid(), due_date(), task_details(), task_uuid(), updated_at()
- invoices:
  - primary key: uuid
  - fields: currency(), customer_uuid(), date(), due_date(), external_id(), uuid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - risk: external mutation; approval required
- update_customer:
  - endpoint: PUT /customers/{{ record.uuid }}
  - required fields: uuid
  - risk: external mutation; approval required

## Security

- read risk: external ChartMogul API read of customer, contact, CRM-task, plan, invoice, and subscription-metrics data
- write risk: external mutation of ChartMogul customer records; approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect chartmogul
```

### Inspect as structured JSON

```bash
pm connectors inspect chartmogul --json
```

## Agent Rules

- Run pm connectors inspect chartmogul before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
