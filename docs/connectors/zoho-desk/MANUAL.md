# pm connectors inspect zoho-desk

```text
NAME
  pm connectors inspect zoho-desk - Zoho Desk connector manual

SYNOPSIS
  pm connectors inspect zoho-desk
  pm connectors inspect zoho-desk --json
  pm credentials add <name> --connector zoho-desk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoho Desk tickets, contacts, and accounts through the Zoho Desk REST API.

ICON
  id: simple-icons-zoho-desk
  asset: icons/simple-icons/zoho-desk.svg
  title: Zoho
  simple_icon_slug: zoho
  simple_icon_hex: E42527
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Zoho
  match: curated-alias
  matched_by: zoho

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  org_id
  page_size
  access_token (secret) (required)

ETL STREAMS
  tickets:
    primary key: id
    cursor: updated_at
    fields: channel(string), createdTime(string), email(string), id(string), modifiedTime(string), name(string), priority(string), status(string), subject(string), ticketNumber(string), updated_at(string)
  contacts:
    primary key: id
    cursor: updated_at
    fields: accountId(string), createdTime(string), email(string), firstName(string), id(string), lastName(string), modifiedTime(string), name(string), phone(string), updated_at(string)
  accounts:
    primary key: id
    cursor: updated_at
    fields: accountName(string), createdTime(string), id(string), modifiedTime(string), name(string), phone(string), updated_at(string), website(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Zoho Desk API read of support ticket and contact data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoho-desk

  # Inspect as structured JSON
  pm connectors inspect zoho-desk --json

AGENT WORKFLOW
  - Run pm connectors inspect zoho-desk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
