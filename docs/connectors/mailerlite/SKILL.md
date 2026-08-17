---
name: pm-mailerlite
description: MailerLite connector knowledge and safe action guide.
---

# pm-mailerlite

## Purpose

Reads MailerLite subscribers, campaigns, groups, segments, and automations through the MailerLite v2 REST API.

## Icon

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
- api_token (secret)

## ETL Streams

- subscribers:
  - primary key: id
  - cursor: updated_at
  - fields: click_rate(), clicks_count(), created_at(), email(), fields(), id(), ip_address(), open_rate(), opens_count(), sent(), source(), status(), subscribed_at(), unsubscribed_at(), updated_at()
- campaigns:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(), created_at(), finished_at(), id(), is_stopped(), name(), scheduled_for(), started_at(), stats(), status(), type(), updated_at()
- groups:
  - primary key: id
  - cursor: created_at
  - fields: active_count(), click_rate(), clicks_count(), created_at(), id(), name(), open_rate(), opens_count(), sent_count(), unsubscribed_count()
- segments:
  - primary key: id
  - cursor: created_at
  - fields: click_rate(), created_at(), id(), name(), open_rate(), total()
- automations:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), enabled(), id(), name(), stats(), status(), steps(), trigger_data()

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
