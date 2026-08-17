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
  items_per_page
  mode
  received_after
  server
  username
  password (secret)

ETL STREAMS
  servers:
    primary key: id
    fields: id(), messages(), name(), users()
  messages:
    primary key: id
    cursor: received
    fields: bcc(), cc(), from(), id(), received(), server(), subject(), to(), type()
  transactions:
    primary key: timestamp
    cursor: timestamp
    fields: email(), previews(), sms(), timestamp()

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
