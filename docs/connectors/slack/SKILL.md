---
name: pm-slack
description: Slack connector knowledge and safe action guide.
---

# pm-slack

## Purpose

Reads Slack workspace users, channels, and channel messages through the Slack Web API. Read-only.

## Icon

- asset: icons/slack.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://api.slack.com/changelog

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- channel_id
- max_pages
- page_size
- access_token (secret)
- api_token (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: deleted(), display_name(), email(), id(), is_admin(), is_bot(), name(), real_name(), team_id(), tz(), updated()
- channels:
  - primary key: id
  - fields: created(), creator(), id(), is_archived(), is_channel(), is_general(), is_group(), is_private(), name(), num_members(), purpose(), topic()
- channel_messages:
  - primary key: ts
  - fields: reply_count(), subtype(), team(), text(), thread_ts(), ts(), type(), user()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Slack Web API read of workspace members/channels/channel message history
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect slack
```

### Inspect as structured JSON

```bash
pm connectors inspect slack --json
```

## Agent Rules

- Run pm connectors inspect slack before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
