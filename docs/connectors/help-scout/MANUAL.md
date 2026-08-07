# pm connectors inspect help-scout

```text
NAME
  pm connectors inspect help-scout - Help Scout connector manual

SYNOPSIS
  pm connectors inspect help-scout
  pm connectors inspect help-scout --json
  pm credentials add <name> --connector help-scout [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Help Scout conversations, customers, mailboxes, and users through the Mailbox API using OAuth2 client-credentials authentication.

ICON
  id: simple-icons-helpscout
  asset: icons/simple-icons/helpscout.svg
  title: Help Scout
  simple_icon_slug: helpscout
  simple_icon_hex: 1292EE
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Help%20Scout
  match: exact-name-or-slug
  matched_by: help-scout

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: userUpdatedAt
    fields: assigneeId(integer), closedAt(string), createdAt(string), folderId(integer), id(integer), mailboxId(integer), number(integer), preview(string), state(string), status(string), subject(string), threads(integer), type(string), userUpdatedAt(string)
  customers:
    primary key: id
    cursor: updatedAt
    fields: age(string), createdAt(string), firstName(string), gender(string), id(integer), jobTitle(string), lastName(string), organization(string), photoUrl(string), updatedAt(string)
  mailboxes:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), email(string), id(integer), name(string), slug(string), updatedAt(string)
  users:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), email(string), firstName(string), id(integer), jobTitle(string), lastName(string), role(string), timezone(string), type(string), updatedAt(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Help Scout API read of conversation, customer, mailbox, and user data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect help-scout

  # Inspect as structured JSON
  pm connectors inspect help-scout --json

AGENT WORKFLOW
  - Run pm connectors inspect help-scout before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
