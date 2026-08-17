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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  guild_id
  mode
  bot_token (secret)

ETL STREAMS
  guilds:
    primary key: id
    fields: approximate_member_count(), approximate_presence_count(), description(), icon(), id(), name(), owner_id(), preferred_locale(), premium_tier()
  channels:
    primary key: id
    fields: guild_id(), id(), name(), nsfw(), parent_id(), position(), topic(), type()
  roles:
    primary key: id
    fields: color(), hoist(), id(), managed(), mentionable(), name(), permissions(), position()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
