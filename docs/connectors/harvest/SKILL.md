---
name: pm-harvest
description: Harvest connector knowledge and safe action guide.
---

# pm-harvest

## Purpose

Reads Harvest clients, contacts, company settings, projects, tasks, task assignments, users, time entries, invoices, estimates, expenses, item categories, expense categories, and roles through the Harvest v2 REST API.

## Icon

- id: harvest
- asset: icons/harvest.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.harveststatus.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id (required)
- base_url
- mode
- page_size
- start_date
- api_token (secret) (required)

## ETL Streams

- clients:
  - primary key: id
  - cursor: updated_at
  - fields: address(string), created_at(string), currency(string), id(integer), is_active(boolean), name(string), statement_key(string), updated_at(string)
- projects:
  - primary key: id
  - cursor: updated_at
  - fields: budget(number), client_id(integer), client_name(string), code(string), created_at(string), id(integer), is_active(boolean), is_billable(boolean), name(string), updated_at(string)
- tasks:
  - primary key: id
  - cursor: updated_at
  - fields: billable_by_default(boolean), created_at(string), default_hourly_rate(number), id(integer), is_active(boolean), is_default(boolean), name(string), updated_at(string)
- users:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), email(string), first_name(string), id(integer), is_active(boolean), is_admin(boolean), last_name(string), timezone(string), updated_at(string)
- time_entries:
  - primary key: id
  - cursor: updated_at
  - fields: billable(boolean), client_id(integer), created_at(string), hours(number), id(integer), is_billed(boolean), is_running(boolean), notes(string), project_id(integer), spent_date(string), task_id(integer), updated_at(string), user_id(integer)
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: client(object), client_id(integer), client_name(string), created_at(string), email(string), fax(string), first_name(string), id(integer), invoice_recipient_status(string), last_name(string), phone_mobile(string), phone_office(string), title(string), updated_at(string)
- company:
  - primary key: full_domain
  - fields: approval_feature(boolean), base_uri(string), clock(string), color_scheme(string), currency_code_display(string), currency_symbol_display(string), date_format(string), decimal_symbol(string), estimate_feature(boolean), expense_feature(boolean), full_domain(string), invoice_feature(boolean), is_active(boolean), name(string), plan_type(string), thousands_separator(string), time_format(string), wants_timestamp_timers(boolean), week_start_day(string), weekly_capacity(integer)
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: amount(number), client(object), client_id(integer), client_key(string), client_name(string), closed_at(string), created_at(string), creator(object), creator_id(integer), creator_name(string), currency(string), discount(number), discount_amount(number), due_amount(number), due_date(string), id(integer), issue_date(string), line_items(array), notes(string), number(string), paid_at(string), paid_date(string), payment_options(array), payment_term(string), period_end(string), period_start(string), purchase_order(string), recurring_invoice_id(integer), sent_at(string), state(string), subject(string), tax(number), tax2(number), tax2_amount(number), tax_amount(number), updated_at(string)
- estimates:
  - primary key: id
  - cursor: updated_at
  - fields: accepted_at(string), amount(number), client(object), client_id(integer), client_key(string), client_name(string), created_at(string), creator(object), creator_id(integer), creator_name(string), currency(string), declined_at(string), discount(number), discount_amount(number), id(integer), issue_date(string), line_items(array), notes(string), number(string), purchase_order(string), sent_at(string), state(string), subject(string), tax(number), tax2(number), tax2_amount(number), tax_amount(number), updated_at(string)
- expenses:
  - primary key: id
  - cursor: updated_at
  - fields: approval_status(string), billable(boolean), client(object), client_id(integer), created_at(string), expense_category(object), expense_category_id(integer), id(integer), invoice(object), invoice_id(integer), is_billed(boolean), is_closed(boolean), is_locked(boolean), locked_reason(string), notes(string), project(object), project_id(integer), receipt(object), spent_date(string), total_cost(number), units(number), updated_at(string), user(object), user_id(integer)
- invoice_item_categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string), use_as_expense(boolean), use_as_service(boolean)
- estimate_item_categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- expense_categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), is_active(boolean), name(string), unit_name(string), unit_price(number), updated_at(string)
- roles:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string), user_ids(array)
- task_assignments:
  - primary key: id
  - cursor: updated_at
  - fields: billable(boolean), budget(number), created_at(string), hourly_rate(number), id(integer), is_active(boolean), project(object), project_id(integer), task(object), task_id(integer), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Harvest API read of Harvest business, time, project, invoice, estimate, expense, role, and category metadata
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect harvest
```

### Inspect as structured JSON

```bash
pm connectors inspect harvest --json
```

## Agent Rules

- Run pm connectors inspect harvest before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
