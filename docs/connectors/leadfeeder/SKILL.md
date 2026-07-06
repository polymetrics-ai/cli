---
name: pm-leadfeeder
description: Leadfeeder connector knowledge and safe action guide.
---

# pm-leadfeeder

## Purpose

Reads Leadfeeder accounts and their leads, visits, and custom feeds through the Leadfeeder JSON:API.

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

- account_id
- base_url
- end_date
- mode
- start_date
- api_token (secret)

## ETL Streams

- accounts:
  - primary key: id
  - fields: currency(), id(), industry(), name(), status(), time_zone(), type()
- leads:
  - primary key: id
  - cursor: last_visit_date
  - fields: city(), country(), employee_count(), first_visit_date(), id(), industry(), last_visit_date(), name(), quality(), type(), visits(), website()
- visits:
  - primary key: id
  - cursor: visit_date
  - fields: ended_at(), hostname(), id(), pageviews(), referring_url(), source(), started_at(), type(), visit_date(), visit_length()
- custom_feeds:
  - primary key: id
  - fields: id(), name(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Leadfeeder API read of account, lead, and visit data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect leadfeeder
```

### Inspect as structured JSON

```bash
pm connectors inspect leadfeeder --json
```

## Agent Rules

- Run pm connectors inspect leadfeeder before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
