---
name: pm-jobnimbus
description: JobNimbus connector knowledge and safe action guide.
---

# pm-jobnimbus

## Purpose

Reads JobNimbus CRM contacts, jobs, tasks, activities, and files through the JobNimbus REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: jnid
  - cursor: date_updated
  - fields: address_line1(), city(), company(), country_name(), date_created(), date_updated(), display_name(), email(), first_name(), home_phone(), is_active(), is_archived(), jnid(), last_name(), mobile_phone(), record_type_name(), sales_rep_name(), state_text(), status_name(), type(), work_phone(), zip()
- jobs:
  - primary key: jnid
  - cursor: date_updated
  - fields: address_line1(), approved_estimate_total(), approved_invoice_total(), city(), customer(), date_created(), date_status_change(), date_updated(), is_active(), is_archived(), jnid(), name(), number(), record_type_name(), sales_rep_name(), state_text(), status_name(), type(), zip()
- tasks:
  - primary key: jnid
  - cursor: date_updated
  - fields: customer(), date_created(), date_end(), date_start(), date_updated(), description(), is_active(), is_archived(), is_completed(), jnid(), number(), priority(), record_type_name(), title(), type()
- activities:
  - primary key: jnid
  - cursor: date_updated
  - fields: created_by_name(), customer(), date_created(), date_updated(), is_active(), is_archived(), is_status_change(), jnid(), note(), record_type_name(), sales_rep_name(), source(), type()
- files:
  - primary key: jnid
  - cursor: date_updated
  - fields: content_type(), created_by_name(), customer(), date_created(), date_file_created(), date_updated(), description(), filename(), is_active(), jnid(), md5(), record_type_name(), size(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external JobNimbus API read of CRM contact, job, task, activity, and file data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect jobnimbus
```

### Inspect as structured JSON

```bash
pm connectors inspect jobnimbus --json
```

## Agent Rules

- Run pm connectors inspect jobnimbus before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
