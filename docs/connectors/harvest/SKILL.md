---
name: pm-harvest
description: Harvest connector knowledge and safe action guide.
---

# pm-harvest

## Purpose

Reads Harvest clients, contacts, company settings, projects, tasks, task assignments, users, time entries, invoices, estimates, expenses, item categories, expense categories, and roles through the Harvest v2 REST API.

## Icon

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

- account_id
- base_url
- mode
- page_size
- start_date
- api_token (secret)

## ETL Streams

- clients:
  - primary key: id
  - cursor: updated_at
  - fields: address(), created_at(), currency(), id(), is_active(), name(), statement_key(), updated_at()
- projects:
  - primary key: id
  - cursor: updated_at
  - fields: budget(), client_id(), client_name(), code(), created_at(), id(), is_active(), is_billable(), name(), updated_at()
- tasks:
  - primary key: id
  - cursor: updated_at
  - fields: billable_by_default(), created_at(), default_hourly_rate(), id(), is_active(), is_default(), name(), updated_at()
- users:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), email(), first_name(), id(), is_active(), is_admin(), last_name(), timezone(), updated_at()
- time_entries:
  - primary key: id
  - cursor: updated_at
  - fields: billable(), client_id(), created_at(), hours(), id(), is_billed(), is_running(), notes(), project_id(), spent_date(), task_id(), updated_at(), user_id()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: client(), client_id(), client_name(), created_at(), email(), fax(), first_name(), id(), invoice_recipient_status(), last_name(), phone_mobile(), phone_office(), title(), updated_at()
- company:
  - primary key: full_domain
  - fields: approval_feature(), base_uri(), clock(), color_scheme(), currency_code_display(), currency_symbol_display(), date_format(), decimal_symbol(), estimate_feature(), expense_feature(), full_domain(), invoice_feature(), is_active(), name(), plan_type(), thousands_separator(), time_format(), wants_timestamp_timers(), week_start_day(), weekly_capacity()
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: amount(), client(), client_id(), client_key(), client_name(), closed_at(), created_at(), creator(), creator_id(), creator_name(), currency(), discount(), discount_amount(), due_amount(), due_date(), id(), issue_date(), line_items(), notes(), number(), paid_at(), paid_date(), payment_options(), payment_term(), period_end(), period_start(), purchase_order(), recurring_invoice_id(), sent_at(), state(), subject(), tax(), tax2(), tax2_amount(), tax_amount(), updated_at()
- estimates:
  - primary key: id
  - cursor: updated_at
  - fields: accepted_at(), amount(), client(), client_id(), client_key(), client_name(), created_at(), creator(), creator_id(), creator_name(), currency(), declined_at(), discount(), discount_amount(), id(), issue_date(), line_items(), notes(), number(), purchase_order(), sent_at(), state(), subject(), tax(), tax2(), tax2_amount(), tax_amount(), updated_at()
- expenses:
  - primary key: id
  - cursor: updated_at
  - fields: approval_status(), billable(), client(), client_id(), created_at(), expense_category(), expense_category_id(), id(), invoice(), invoice_id(), is_billed(), is_closed(), is_locked(), locked_reason(), notes(), project(), project_id(), receipt(), spent_date(), total_cost(), units(), updated_at(), user(), user_id()
- invoice_item_categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), updated_at(), use_as_expense(), use_as_service()
- estimate_item_categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), updated_at()
- expense_categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), is_active(), name(), unit_name(), unit_price(), updated_at()
- roles:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), updated_at(), user_ids()
- task_assignments:
  - primary key: id
  - cursor: updated_at
  - fields: billable(), budget(), created_at(), hourly_rate(), id(), is_active(), project(), project_id(), task(), task_id(), updated_at()

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
