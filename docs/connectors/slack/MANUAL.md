# pm connectors inspect slack

```text
NAME
  pm connectors inspect slack - Slack connector manual

SYNOPSIS
  pm connectors inspect slack
  pm connectors inspect slack --json
  pm credentials add <name> --connector slack [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Slack workspace users, channels, and channel messages through the Slack Web API. Read-only.

ICON
  asset: icons/slack.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.slack.com/changelog

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  channel_id
  max_pages
  page_size
  access_token (secret)
  api_token (secret)

ETL STREAMS
  users:
    primary key: id
    fields: deleted(), display_name(), email(), id(), is_admin(), is_bot(), name(), real_name(), team_id(), tz(), updated()
  channels:
    primary key: id
    fields: created(), creator(), id(), is_archived(), is_channel(), is_general(), is_group(), is_private(), name(), num_members(), purpose(), topic()
  channel_messages:
    primary key: ts
    fields: reply_count(), subtype(), team(), text(), thread_ts(), ts(), type(), user()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Slack Web API read of workspace members/channels/channel message history
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect slack

  # Inspect as structured JSON
  pm connectors inspect slack --json

AGENT WORKFLOW
  - Run pm connectors inspect slack before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
