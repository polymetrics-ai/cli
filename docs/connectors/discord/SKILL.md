---
name: pm-discord
description: Discord connector knowledge and safe action guide.
---

# pm-discord

## Purpose

Reads Discord guild, channel, and role data through the Discord REST API using a bot token. The members stream is out of scope for this migration (see docs.md's Known limits).

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
- guild_id
- mode
- bot_token (secret)

## ETL Streams

- guilds:
  - primary key: id
  - fields: approximate_member_count(), approximate_presence_count(), description(), icon(), id(), name(), owner_id(), preferred_locale(), premium_tier()
- channels:
  - primary key: id
  - fields: guild_id(), id(), name(), nsfw(), parent_id(), position(), topic(), type()
- roles:
  - primary key: id
  - fields: color(), hoist(), id(), managed(), mentionable(), name(), permissions(), position()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Discord API read of guild, channel, and role data
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect discord
```

### Inspect as structured JSON

```bash
pm connectors inspect discord --json
```

## Agent Rules

- Run pm connectors inspect discord before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
