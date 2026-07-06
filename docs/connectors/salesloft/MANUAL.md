# pm connectors inspect salesloft

```text
NAME
  pm connectors inspect salesloft - Salesloft connector manual

SYNOPSIS
  pm connectors inspect salesloft
  pm connectors inspect salesloft --json
  pm credentials add <name> --connector salesloft [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Salesloft people, accounts, cadences, users, and emails through the Salesloft REST API v2.

ICON
  asset: icons/salesloft.svg
  source: official
  review_status: official_verified
  review_url: https://developers.salesloft.com/docs/api/

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
  start_date
  token_url
  access_token (secret)
  api_key (secret)
  client_id (secret)
  client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  people:
    primary key: id
    cursor: updated_at
    fields: account_id(), created_at(), display_name(), do_not_contact(), email_address(), first_name(), id(), last_name(), owner_id(), person_company_name(), phone(), title(), updated_at()
  accounts:
    primary key: id
    cursor: updated_at
    fields: archived_at(), city(), company_type(), country(), created_at(), domain(), id(), industry(), name(), owner_id(), phone(), updated_at(), website()
  cadences:
    primary key: id
    cursor: updated_at
    fields: archived_at(), created_at(), id(), name(), remove_bounces_enabled(), remove_replies_enabled(), shared(), team_cadence(), updated_at()
  users:
    primary key: id
    cursor: updated_at
    fields: active(), created_at(), email(), first_name(), guid(), id(), last_name(), name(), time_zone(), updated_at()
  emails:
    primary key: id
    cursor: updated_at
    fields: bounced(), click_tracking(), created_at(), id(), sent_at(), status(), subject(), updated_at(), view_tracking()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Salesloft API read of people, accounts, cadences, users, and email data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect salesloft

  # Inspect as structured JSON
  pm connectors inspect salesloft --json

AGENT WORKFLOW
  - Run pm connectors inspect salesloft before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
