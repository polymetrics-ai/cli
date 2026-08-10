# pm connectors inspect rocket-chat

```text
NAME
  pm connectors inspect rocket-chat - Rocket.Chat connector manual

SYNOPSIS
  pm connectors inspect rocket-chat
  pm connectors inspect rocket-chat --json
  pm credentials add <name> --connector rocket-chat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Rocket.Chat users, public channels, private groups, direct messages, and rooms through the REST API.

ICON
  id: rocket-chat
  asset: icons/rocket-chat.svg
  source: official
  review_status: official_verified
  review_url: https://developer.rocket.chat/apidocs

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  fields
  mode
  query
  room_id
  updated_since
  auth_token (secret) (required)
  user_id (secret) (required)

ETL STREAMS
  users:
    primary key: id
    cursor: updated_at
    fields: emails(array), id(string), name(string), status(string), stream(string), updated_at(string), username(string)
  channels:
    primary key: id
    cursor: updated_at
    fields: fname(string), id(string), msgs(integer), name(string), stream(string), updated_at(string)
  groups:
    primary key: id
    cursor: updated_at
    fields: fname(string), id(string), msgs(integer), name(string), stream(string), updated_at(string)
  direct_messages:
    primary key: id
    cursor: updated_at
    fields: id(string), msgs(integer), stream(string), updated_at(string), usernames(array)
  rooms:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), stream(string), type(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Rocket.Chat API read of workspace users, rooms, and messages metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect rocket-chat

  # Inspect as structured JSON
  pm connectors inspect rocket-chat --json

AGENT WORKFLOW
  - Run pm connectors inspect rocket-chat before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
