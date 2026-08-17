# pm connectors inspect zendesk-chat

```text
NAME
  pm connectors inspect zendesk-chat - Zendesk Chat connector manual

SYNOPSIS
  pm connectors inspect zendesk-chat
  pm connectors inspect zendesk-chat --json
  pm credentials add <name> --connector zendesk-chat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zendesk Chat agents, chats, departments, shortcuts, and triggers through the Zendesk Chat REST API v2.

ICON
  asset: icons/zendesk-chat.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://support.zendesk.com/hc/en-us/sections/4405298889242-Developer-updates

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  access_token (secret)

ETL STREAMS
  agents:
    primary key: id
    fields: create_date(), display_name(), email(), enabled(), first_name(), id(), last_login(), last_name(), role_id()
  chats:
    primary key: id
    cursor: timestamp
    fields: comment(), department_id(), duration(), id(), rating(), session(), timestamp(), type(), visitor()
  departments:
    primary key: id
    fields: description(), enabled(), id(), members(), name(), settings()
  shortcuts:
    primary key: id
    fields: id(), message(), name(), options(), scope(), tags()
  triggers:
    primary key: id
    fields: definition(), description(), enabled(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Zendesk Chat API read of agent, chat, and configuration data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zendesk-chat

  # Inspect as structured JSON
  pm connectors inspect zendesk-chat --json

AGENT WORKFLOW
  - Run pm connectors inspect zendesk-chat before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
