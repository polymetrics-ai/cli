# pm connectors inspect freshdesk

```text
NAME
  pm connectors inspect freshdesk - Freshdesk connector manual

SYNOPSIS
  pm connectors inspect freshdesk
  pm connectors inspect freshdesk --json
  pm credentials add <name> --connector freshdesk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freshdesk tickets, contacts, companies, agents, and groups through the Freshdesk REST API v2.

ICON
  id: freshdesk
  asset: icons/freshdesk.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.freshdesk.com/api/#change_log

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  max_pages
  mode
  page_size
  start_date
  api_key (secret) (required)

ETL STREAMS
  tickets:
    primary key: id
    cursor: updated_at
    fields: company_id(integer), created_at(string), due_by(string), group_id(integer), id(integer), priority(integer), requester_id(integer), responder_id(integer), source(integer), spam(boolean), status(integer), subject(string), type(string), updated_at(string)
  contacts:
    primary key: id
    cursor: updated_at
    fields: active(boolean), company_id(integer), created_at(string), email(string), id(integer), mobile(string), name(string), phone(string), updated_at(string)
  companies:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(integer), name(string), note(string), updated_at(string)
  agents:
    primary key: id
    cursor: updated_at
    fields: available(boolean), created_at(string), id(integer), occasional(boolean), ticket_scope(integer), updated_at(string)
  groups:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(integer), name(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Freshdesk API read of support tickets, contacts, companies, agents, and groups
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect freshdesk

  # Inspect as structured JSON
  pm connectors inspect freshdesk --json

AGENT WORKFLOW
  - Run pm connectors inspect freshdesk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
