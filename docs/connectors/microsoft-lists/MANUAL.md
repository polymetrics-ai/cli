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
  list_id
  login_base_url
  max_pages
  mode
  page_size
  scope
  site_id
  token_url
  client_id (secret)
  client_secret (secret)
  tenant_id (secret)

ETL STREAMS
  lists:
    primary key: id
    cursor: last_modified_date_time
    fields: created_date_time(), description(), display_name(), etag(), id(), last_modified_date_time(), list_template(), name(), web_url()
  list_items:
    primary key: id
    cursor: last_modified_date_time
    fields: content_type_id(), created_date_time(), etag(), fields(), id(), last_modified_date_time(), web_url()
  columns:
    primary key: id
    fields: column_group(), description(), display_name(), hidden(), id(), indexed(), name(), read_only(), required()
  content_types:
    primary key: id
    fields: description(), group(), hidden(), id(), name(), read_only(), sealed()

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
