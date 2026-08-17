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
  private_key (secret)

ETL STREAMS
  domains:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), is_disabled(), name(), smtp_login(), spam_action(), state(), type(), wildcard()
  events:
    primary key: id
    cursor: timestamp
    fields: event(), id(), log_level(), message_id(), reason(), recipient(), timestamp()
  mailing_lists:
    primary key: address
    cursor: created_at
    fields: access_level(), address(), created_at(), description(), members_count(), name()
  tags:
    primary key: tag
    fields: description(), first_seen(), last_seen(), tag()

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
