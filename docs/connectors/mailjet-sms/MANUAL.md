# pm connectors inspect mailjet-sms

```text
NAME
  pm connectors inspect mailjet-sms - Mailjet SMS connector manual

SYNOPSIS
  pm connectors inspect mailjet-sms
  pm connectors inspect mailjet-sms --json
  pm credentials add <name> --connector mailjet-sms [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailjet SMS messages, message counts, and export job status; writes SMS send and export-request actions.

ICON
  id: mailjetsms
  asset: icons/mailjetsms.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://dev.mailjet.com/sms/reference/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  end_date
  export_job_id
  mode
  recipient
  sms_id
  sms_ids
  start_date
  status_code
  token (secret) (required)

ETL STREAMS
  sms:
    primary key: ID
    cursor: CreationTS
    fields: CreationTS(integer), From(string), ID(string), MessageId(string), SMSCount(integer), SentTS(integer), To(string), cost_currency(string), cost_value(number), status_code(integer), status_description(string), status_name(string)
  sms_count:
    fields: Count(integer)
  sms_message:
    primary key: MessageID
    cursor: CreationTS
    fields: CreationTS(integer), From(string), MessageID(string), SMSCount(integer), SentTS(integer), To(string), cost_currency(string), cost_value(number), status_code(integer), status_description(string), status_name(string)
  sms_export:
    primary key: ID
    cursor: CreationTS
    fields: CreationTS(integer), ExpirationTS(integer), ID(integer), URL(string), status_code(integer), status_description(string), status_name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  send_sms:
    endpoint: POST /sms-send
    required fields: From, To, Text
    risk: external mutation; sends an SMS message; approval required
  request_sms_export:
    endpoint: POST /sms/export
    required fields: FromTS, ToTS
    risk: external mutation; creates an asynchronous SMS export job; approval required

SECURITY
  read risk: external Mailjet SMS API read of outbound SMS message data
  write risk: external Mailjet SMS API mutation; may send SMS messages or request asynchronous SMS exports
  approval: required for all write actions
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailjet-sms

  # Inspect as structured JSON
  pm connectors inspect mailjet-sms --json

AGENT WORKFLOW
  - Run pm connectors inspect mailjet-sms before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
