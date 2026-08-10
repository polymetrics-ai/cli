# pm connectors inspect microsoft-lists

```text
NAME
  pm connectors inspect microsoft-lists - Microsoft Lists connector manual

SYNOPSIS
  pm connectors inspect microsoft-lists
  pm connectors inspect microsoft-lists --json
  pm credentials add <name> --connector microsoft-lists [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SharePoint/Microsoft Lists, list items, columns, and content types from a site through the Microsoft Graph API using an OAuth2 client-credentials grant. Read-only.

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
  list_id
  login_base_url
  max_pages
  mode
  page_size
  scope
  site_id (required)
  token_url
  client_id (secret)
  client_secret (secret)
  tenant_id (secret)

ETL STREAMS
  lists:
    primary key: id
    cursor: last_modified_date_time
    fields: created_date_time(string), description(string), display_name(string), etag(string), id(string), last_modified_date_time(string), list_template(string), name(string), web_url(string)
  list_items:
    primary key: id
    cursor: last_modified_date_time
    fields: content_type_id(string), created_date_time(string), etag(string), fields(object), id(string), last_modified_date_time(string), web_url(string)
  columns:
    primary key: id
    fields: column_group(string), description(string), display_name(string), hidden(boolean), id(string), indexed(boolean), name(string), read_only(boolean), required(boolean)
  content_types:
    primary key: id
    fields: description(string), group(string), hidden(boolean), id(string), name(string), read_only(boolean), sealed(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Microsoft Graph API read of a SharePoint site's lists/list items/columns/content types
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect microsoft-lists

  # Inspect as structured JSON
  pm connectors inspect microsoft-lists --json

AGENT WORKFLOW
  - Run pm connectors inspect microsoft-lists before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
