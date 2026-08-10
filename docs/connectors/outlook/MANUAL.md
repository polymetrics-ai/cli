# pm connectors inspect outlook

```text
NAME
  pm connectors inspect outlook - Outlook connector manual

SYNOPSIS
  pm connectors inspect outlook
  pm connectors inspect outlook --json
  pm credentials add <name> --connector outlook [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Outlook messages, mail folders, and calendar events through Microsoft Graph using an OAuth 2.0 refresh-token grant.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

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
  scope
  tenant_id
  token_url
  client_id (secret) (required)
  client_secret (secret) (required)
  refresh_token (secret) (required)

ETL STREAMS
  messages:
    primary key: id
    cursor: received_date_time
    fields: id(string), last_modified_date_time(string), received_date_time(string), subject(string), web_link(string)
  mail_folders:
    primary key: id
    fields: display_name(string), id(string), total_item_count(integer), unread_item_count(integer)
  events:
    primary key: id
    cursor: last_modified_date_time
    fields: created_date_time(string), id(string), last_modified_date_time(string), subject(string), web_link(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Microsoft Graph API read of the authenticated mailbox's messages, mail folders, and calendar events
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect outlook

  # Inspect as structured JSON
  pm connectors inspect outlook --json

AGENT WORKFLOW
  - Run pm connectors inspect outlook before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
