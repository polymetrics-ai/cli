---
name: pm-discord
description: Discord connector knowledge and safe action guide.
---

# pm-discord

## Purpose

Reads Discord guild, channel, and role data through the Discord REST API using a bot token. The members stream is out of scope for this migration (see docs.md's Known limits).

## Icon

- id: simple-icons-discord
- asset: icons/simple-icons/discord.svg
- title: Discord
- simple_icon_slug: discord
- simple_icon_hex: 5865F2
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Discord
- match: exact-name-or-slug
- matched_by: discord

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- guild_id (required)
- mode
- bot_token (secret) (required)

## ETL Streams

- guilds:
  - primary key: id
  - fields: approximate_member_count(integer), approximate_presence_count(integer), description(string), icon(string), id(string), name(string), owner_id(string), preferred_locale(string), premium_tier(integer)
- channels:
  - primary key: id
  - fields: guild_id(string), id(string), name(string), nsfw(boolean), parent_id(string), position(integer), topic(string), type(integer)
- roles:
  - primary key: id
  - fields: color(integer), hoist(boolean), id(string), managed(boolean), mentionable(boolean), name(string), permissions(string), position(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
