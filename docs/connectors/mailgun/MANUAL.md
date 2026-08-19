# pm connectors inspect mailgun

```text
NAME
  pm connectors inspect mailgun - Mailgun connector manual

SYNOPSIS
  pm connectors inspect mailgun
  pm connectors inspect mailgun --json
  pm credentials add <name> --connector mailgun [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailgun sending domains, email events, mailing lists, and analytics tags through the Mailgun v3 REST API.

ICON
  id: mailgun
  asset: icons/mailgun.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://documentation.mailgun.com/en/latest/api_reference.html

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  domain_name
  mode
  page_size
  private_key (secret) (required)

ETL STREAMS
  domains:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(string), is_disabled(boolean), name(string), smtp_login(string), spam_action(string), state(string), type(string), wildcard(boolean)
  events:
    primary key: id
    cursor: timestamp
    fields: event(string), id(string), log_level(string), message_id(string), reason(string), recipient(string), timestamp(number)
  mailing_lists:
    primary key: address
    cursor: created_at
    fields: access_level(string), address(string), created_at(string), description(string), members_count(integer), name(string)
  tags:
    primary key: tag
    fields: description(string), first_seen(string), last_seen(string), tag(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Mailgun API read of sending-domain, event, mailing-list, and tag data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailgun

  # Inspect as structured JSON
  pm connectors inspect mailgun --json

AGENT WORKFLOW
  - Run pm connectors inspect mailgun before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
