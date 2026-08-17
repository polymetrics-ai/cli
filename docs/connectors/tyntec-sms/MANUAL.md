# pm connectors inspect tyntec-sms

```text
NAME
  pm connectors inspect tyntec-sms - tyntec SMS connector manual

SYNOPSIS
  pm connectors inspect tyntec-sms
  pm connectors inspect tyntec-sms --json
  pm credentials add <name> --connector tyntec-sms [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads tyntec SMS messages, templates, sender IDs, and delivery reports through API list endpoints, and sends approved SMS messages through the Messaging API.

ICON
  asset: icons/tyntec.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.tyntec.com/reference/messaging

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  messages:
    primary key: id
    cursor: created_at
    fields: created_at(), from(), id(), status(), to()
  templates:
    primary key: id
    fields: id(), name()
  sender_ids:
    primary key: id
    fields: id(), name()
  delivery_reports:
    primary key: id
    cursor: created_at
    fields: created_at(), from(), id(), status(), to()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  send_message:
    endpoint: POST sms/v1/messages
    risk: sends a billable SMS message to the recipient phone number and may notify an external user

SECURITY
  read risk: external tyntec SMS API read of messages, templates, sender IDs, and delivery reports
  write risk: sends billable SMS messages to recipient phone numbers; approval required before delivery
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tyntec-sms

  # Inspect as structured JSON
  pm connectors inspect tyntec-sms --json

AGENT WORKFLOW
  - Run pm connectors inspect tyntec-sms before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
