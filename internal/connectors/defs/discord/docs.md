# Overview

Reads Discord guild, channel, and role data through the Discord REST API using a bot token.

Readable streams: `guilds`, `channels`, `roles`.

This connector is read-only; no write actions are declared.

Service API documentation: https://discord.com/developers/docs/reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://discord.com/api/v10`; format `uri`; Discord API
  base URL override for tests or proxies.
- `bot_token` (required, secret, string); Discord bot token, sent as Authorization: Bot <bot_token>.
  Used only for the Authorization header; never logged.
- `guild_id` (required, string); Discord guild (server) id; every stream is scoped to this guild.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `bot_token`.

Default configuration values: `base_url=https://discord.com/api/v10`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Bot` using `secrets.bot_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/@me`.

## Streams notes

Default pagination: single request; no pagination.

- `guilds`: GET `/guilds/{{ config.guild_id }}` - single-object response; records path `.`.
- `channels`: GET `/guilds/{{ config.guild_id }}/channels` - records at response root.
- `roles`: GET `/guilds/{{ config.guild_id }}/roles` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external Discord API read of guild, channel, and role
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
