# pm connectors inspect freshcaller

```text
NAME
  pm connectors inspect freshcaller - Freshcaller connector manual

SYNOPSIS
  pm connectors inspect freshcaller
  pm connectors inspect freshcaller --json
  pm credentials add <name> --connector freshcaller [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freshcaller calls, agents, teams, and phone numbers through the Freshcaller REST API.

ICON
  id: freshcaller
  asset: icons/freshcaller.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.freshcaller.com/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  api_key (secret) (required)

ETL STREAMS
  calls:
    primary key: id
    cursor: call_time
    fields: agent_id(integer), call_time(string), direction(string), duration(integer), id(integer), phone_number(string), status(string)
  agents:
    primary key: id
    fields: email(string), id(integer), name(string), status(string)
  teams:
    primary key: id
    fields: id(integer), name(string)
  numbers:
    primary key: id
    fields: id(integer), name(string), phone_number(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Freshcaller API read of call, agent, team, and phone number data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect freshcaller

  # Inspect as structured JSON
  pm connectors inspect freshcaller --json

AGENT WORKFLOW
  - Run pm connectors inspect freshcaller before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
