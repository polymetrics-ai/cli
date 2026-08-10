---
name: pm-flowlu
description: Flowlu connector knowledge and safe action guide.
---

# pm-flowlu

## Purpose

Reads Flowlu CRM accounts, leads, tasks, projects, invoices, and agile issues through the Flowlu REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- company (required)
- api_key (secret)

## ETL Streams

- accounts:
  - primary key: id
  - cursor: updated_date
  - fields: active(integer), created_date(string), email(string), first_name(string), id(integer), last_name(string), name(string), owner_id(integer), phone(string), type(integer), updated_date(string)
- leads:
  - primary key: id
  - cursor: updated_date
  - fields: active(integer), budget(string), created_date(string), id(integer), name(string), owner_id(integer), pipeline_id(integer), stage_id(integer), title(string), updated_date(string)
- tasks:
  - primary key: id
  - cursor: updated_date
  - fields: created_date(string), deadline(string), description(string), id(integer), name(string), owner_id(integer), priority(integer), responsible_id(integer), updated_date(string), workflow_stage_id(integer)
- projects:
  - primary key: id
  - cursor: updated_date
  - fields: active(integer), created_date(string), description(string), id(integer), manager_id(integer), name(string), owner_id(integer), stage_id(integer), updated_date(string)
- invoices:
  - primary key: id
  - cursor: updated_date
  - fields: created_date(string), currency_id(integer), customer_id(integer), id(integer), invoice_date(string), invoice_number(string), invoice_status(integer), name(string), total_amount(string), updated_date(string)
- agile_issues:
  - primary key: id
  - cursor: updated_date
  - fields: created_date(string), description(string), id(integer), name(string), owner_id(integer), priority(integer), project_id(integer), sprint_id(integer), type(integer), updated_date(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Flowlu CRM read of accounts/leads/tasks/projects/invoices/agile issues
- approval: none; read-only API-key access
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect flowlu
```

### Inspect as structured JSON

```bash
pm connectors inspect flowlu --json
```

## Agent Rules

- Run pm connectors inspect flowlu before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
