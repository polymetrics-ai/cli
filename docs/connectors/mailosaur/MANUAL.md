# pm connectors inspect mailosaur

```text
NAME
  pm connectors inspect mailosaur - Mailosaur connector manual

SYNOPSIS
  pm connectors inspect mailosaur
  pm connectors inspect mailosaur --json
  pm credentials add <name> --connector mailosaur [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailosaur virtual servers, message summaries, and account usage transactions through the Mailosaur REST API.

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
  items_per_page
  mode
  received_after
  server
  username
  password (secret) (required)

ETL STREAMS
  servers:
    primary key: id
    fields: id(string), messages(integer), name(string), users(array)
  messages:
    primary key: id
    cursor: received
    fields: bcc(array), cc(array), from(array), id(string), received(string), server(string), subject(string), to(array), type(string)
  transactions:
    primary key: timestamp
    cursor: timestamp
    fields: email(integer), previews(integer), sms(integer), timestamp(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Mailosaur API read of virtual-server, message-summary, and usage-transaction data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailosaur

  # Inspect as structured JSON
  pm connectors inspect mailosaur --json

AGENT WORKFLOW
  - Run pm connectors inspect mailosaur before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
