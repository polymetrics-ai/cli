# pm connectors inspect discord

```text
NAME
  pm connectors inspect discord - Discord connector manual

SYNOPSIS
  pm connectors inspect discord
  pm connectors inspect discord --json
  pm credentials add <name> --connector discord [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Discord guild, channel, and role data through the Discord REST API using a bot token. The members stream is out of scope for this migration (see docs.md's Known limits).

ICON
  id: simple-icons-discord
  asset: icons/simple-icons/discord.svg
  title: Discord
  simple_icon_slug: discord
  simple_icon_hex: 5865F2
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Discord
  match: exact-name-or-slug
  matched_by: discord

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  guild_id (required)
  mode
  bot_token (secret) (required)

ETL STREAMS
  guilds:
    primary key: id
    fields: approximate_member_count(integer), approximate_presence_count(integer), description(string), icon(string), id(string), name(string), owner_id(string), preferred_locale(string), premium_tier(integer)
  channels:
    primary key: id
    fields: guild_id(string), id(string), name(string), nsfw(boolean), parent_id(string), position(integer), topic(string), type(integer)
  roles:
    primary key: id
    fields: color(integer), hoist(boolean), id(string), managed(boolean), mentionable(boolean), name(string), permissions(string), position(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Discord API read of guild, channel, and role data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect discord

  # Inspect as structured JSON
  pm connectors inspect discord --json

AGENT WORKFLOW
  - Run pm connectors inspect discord before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
