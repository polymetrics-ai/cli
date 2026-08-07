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
  id: salesloft
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
    fields: account_id(integer), created_at(string), display_name(string), do_not_contact(boolean), email_address(string), first_name(string), id(integer), last_name(string), owner_id(integer), person_company_name(string), phone(string), title(string), updated_at(string)
  accounts:
    primary key: id
    cursor: updated_at
    fields: archived_at(string), city(string), company_type(string), country(string), created_at(string), domain(string), id(integer), industry(string), name(string), owner_id(integer), phone(string), updated_at(string), website(string)
  cadences:
    primary key: id
    cursor: updated_at
    fields: archived_at(string), created_at(string), id(integer), name(string), remove_bounces_enabled(boolean), remove_replies_enabled(boolean), shared(boolean), team_cadence(boolean), updated_at(string)
  users:
    primary key: id
    cursor: updated_at
    fields: active(boolean), created_at(string), email(string), first_name(string), guid(string), id(integer), last_name(string), name(string), time_zone(string), updated_at(string)
  emails:
    primary key: id
    cursor: updated_at
    fields: bounced(boolean), click_tracking(boolean), created_at(string), id(integer), sent_at(string), status(string), subject(string), updated_at(string), view_tracking(boolean)

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
