---
name: pm-mailerlite
description: MailerLite connector knowledge and safe action guide.
---

# pm-mailerlite

## Purpose

Reads MailerLite subscribers, campaigns, groups, segments, and automations through the MailerLite v2 REST API.

## Icon

- id: mailerlite
- asset: icons/mailerlite.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.mailerlite.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_token (secret) (required)

## ETL Streams

- subscribers:
  - primary key: id
  - cursor: updated_at
  - fields: click_rate(number), clicks_count(integer), created_at(string), email(string), fields(object), id(string), ip_address(string), open_rate(number), opens_count(integer), sent(integer), source(string), status(string), subscribed_at(string), unsubscribed_at(string), updated_at(string)
- campaigns:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(string), created_at(string), finished_at(string), id(string), is_stopped(boolean), name(string), scheduled_for(string), started_at(string), stats(object), status(string), type(string), updated_at(string)
- groups:
  - primary key: id
  - cursor: created_at
  - fields: active_count(integer), click_rate(object), clicks_count(integer), created_at(string), id(string), name(string), open_rate(object), opens_count(integer), sent_count(integer), unsubscribed_count(integer)
- segments:
  - primary key: id
  - cursor: created_at
  - fields: click_rate(object), created_at(string), id(string), name(string), open_rate(object), total(integer)
- automations:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), enabled(boolean), id(string), name(string), stats(object), status(string), steps(object), trigger_data(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external MailerLite API read of subscriber, campaign, group, segment, and automation data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mailerlite
```

### Inspect as structured JSON

```bash
pm connectors inspect mailerlite --json
```

## Agent Rules

- Run pm connectors inspect mailerlite before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
