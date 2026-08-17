# pm connectors inspect front

```text
NAME
  pm connectors inspect front - Front connector manual

SYNOPSIS
  pm connectors inspect front
  pm connectors inspect front --json
  pm credentials add <name> --connector front [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Front contacts, conversations, inboxes, tags, teammates, and channels through the Front Core REST API.

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
  page_limit
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), id(), is_private(), is_spammer(), name(), updated_at()
  conversations:
    primary key: id
    cursor: last_message_at
    fields: created_at(), id(), is_private(), last_message_at(), status(), subject(), waiting_since()
  inboxes:
    primary key: id
    fields: custom_fields(), id(), is_private(), is_public(), name()
  tags:
    primary key: id
    fields: created_at(), highlight(), id(), is_private(), is_visible_in_conversation_lists(), name(), updated_at()
  teammates:
    primary key: id
    fields: email(), first_name(), id(), is_admin(), is_available(), is_blocked(), last_name(), username()
  channels:
    primary key: id
    fields: address(), id(), is_private(), is_valid(), name(), send_as(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Front API read of contact, conversation, inbox, tag, teammate, and channel data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect front

  # Inspect as structured JSON
  pm connectors inspect front --json

AGENT WORKFLOW
  - Run pm connectors inspect front before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
