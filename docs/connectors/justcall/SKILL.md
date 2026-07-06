---
name: pm-justcall
description: JustCall connector knowledge and safe action guide.
---

# pm-justcall

## Purpose

Reads JustCall users, call logs, SMS, contacts, and phone numbers through the JustCall REST API.

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
- mode
- page_size
- start_date
- api_key_2 (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: available(), created_at(), email(), extension(), id(), last_login_timestamp(), name(), on_call(), role(), timezone()
- calls:
  - primary key: id
  - cursor: call_date
  - fields: agent_email(), agent_id(), agent_name(), call_date(), call_duration(), call_sid(), call_time(), contact_name(), contact_number(), cost_incurred(), id(), justcall_line_name(), justcall_number()
- sms:
  - primary key: id
  - cursor: sms_date
  - fields: agent_email(), agent_id(), agent_name(), contact_name(), contact_number(), cost_incurred(), delivery_status(), direction(), id(), justcall_line_name(), justcall_number(), sms_date(), sms_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external JustCall API read of users, call logs, SMS, contacts, and phone numbers
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect justcall
```

### Inspect as structured JSON

```bash
pm connectors inspect justcall --json
```

## Agent Rules

- Run pm connectors inspect justcall before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
