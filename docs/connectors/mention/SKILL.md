---
name: pm-mention
description: Mention connector knowledge and safe action guide.
---

# pm-mention

## Purpose

Reads Mention app metadata, accounts, alerts, mentions, alert tags, alert shares, alert preferences, and alert tasks from the Mention social listening REST API.

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

- account_id (required)
- alert_id
- base_url
- mode
- api_key (secret) (required)

## ETL Streams

- app_data:
  - fields: actions(object), alert_languages(object), countries(object), days(object), folders(object), integrations(object), languages(object), sources(object), tones(object)
- account_me:
  - primary key: id
  - fields: created_at(string), id(string), language(string), name(string), permission(string), timezone(string)
- account:
  - primary key: id
  - fields: created_at(string), id(string), language(string), name(string), permission(string), timezone(string)
- alert:
  - primary key: id
  - fields: countries(array), created_at(string), description(string), id(string), languages(array), name(string), query(object), sources(array), updated_at(string)
- mention:
  - primary key: id
  - fields: created_at(string), description(string), favorite(boolean), id(string), language(string), published_at(string), source_name(string), source_type(string), title(string), tone(number), url(string)
- alert_tag:
  - primary key: id
  - fields: color(string), id(string), name(string)
- alert_share:
  - primary key: id
  - fields: created_at(string), email(string), id(string), permission(string), updated_at(string)
- alert_preferences:
  - fields: frequency(string), notification_frequency(string), send_email(boolean), send_push(boolean), shared(boolean)
- alert_task:
  - primary key: id
  - fields: created_at(string), description(string), id(string), mention(object), state(string), title(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Mention API read of app metadata, account, alert, mention, tag, share, preference, and task data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mention
```

### Inspect as structured JSON

```bash
pm connectors inspect mention --json
```

## Agent Rules

- Run pm connectors inspect mention before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
