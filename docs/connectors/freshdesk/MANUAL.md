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

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  fixture: Fixture-backed conformance mode; no credentials required.
    supports: read=true write=false
  api_key: Live Freshdesk API key via HTTP Basic auth (api_key:X).
    secrets: api_key
    supports: read=true write=false

CONFIGURATION
  domain (required): Freshdesk domain, e.g. acme.freshdesk.com. Used to build the https://<domain>/api/v2 base URL.
  base_url: Freshdesk base URL override for tests or proxies; /api/v2 is appended automatically.
  start_date: RFC3339 lower bound; for tickets, only objects updated at or after this time are read (updated_since).
  page_size default=100: Records per page (1-100).
  max_pages default=0: Maximum pages; use 0, all, or unlimited to exhaust the stream.
  mode: Runtime mode: live (default) or fixture for credential-free conformance.
  api_key (secret) (required): Freshdesk API key. Sent as the HTTP Basic username (password is the literal X); never logged.

ETL STREAMS
  tickets: Freshdesk support tickets.
    primary key: id
    cursor: updated_at
    fields: id(integer), subject(string), type(string), status(integer), priority(integer), source(integer), requester_id(integer), responder_id(integer), group_id(integer), company_id(integer), spam(boolean), due_by(timestamp), created_at(timestamp), updated_at(timestamp)
  contacts: Freshdesk contacts (requesters).
    primary key: id
    cursor: updated_at
    fields: id(integer), name(string), email(string), phone(string), mobile(string), company_id(integer), active(boolean), created_at(timestamp), updated_at(timestamp)
  companies: Freshdesk companies.
    primary key: id
    cursor: updated_at
    fields: id(integer), name(string), description(string), note(string), created_at(timestamp), updated_at(timestamp)
  agents: Freshdesk agents.
    primary key: id
    cursor: updated_at
    fields: id(integer), available(boolean), occasional(boolean), ticket_scope(integer), created_at(timestamp), updated_at(timestamp)
  groups: Freshdesk agent groups.
    primary key: id
    cursor: updated_at
    fields: id(integer), name(string), description(string), created_at(timestamp), updated_at(timestamp)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
  Source modes: full_refresh, incremental

PAGINATION
  type: link_header
  page size field: page_size
  page limit field: max_pages
  default limit: 0

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
