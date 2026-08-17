# pm connectors inspect justcall

```text
NAME
  pm connectors inspect justcall - JustCall connector manual

SYNOPSIS
  pm connectors inspect justcall
  pm connectors inspect justcall --json
  pm credentials add <name> --connector justcall [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads JustCall users, call logs, SMS, contacts, and phone numbers through the JustCall REST API.

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
  mode
  page_size
  start_date
  api_key_2 (secret)

ETL STREAMS
  users:
    primary key: id
    fields: available(), created_at(), email(), extension(), id(), last_login_timestamp(), name(), on_call(), role(), timezone()
  calls:
    primary key: id
    cursor: call_date
    fields: agent_email(), agent_id(), agent_name(), call_date(), call_duration(), call_sid(), call_time(), contact_name(), contact_number(), cost_incurred(), id(), justcall_line_name(), justcall_number()
  sms:
    primary key: id
    cursor: sms_date
    fields: agent_email(), agent_id(), agent_name(), contact_name(), contact_number(), cost_incurred(), delivery_status(), direction(), id(), justcall_line_name(), justcall_number(), sms_date(), sms_time()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external JustCall API read of users, call logs, SMS, contacts, and phone numbers
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect justcall

  # Inspect as structured JSON
  pm connectors inspect justcall --json

AGENT WORKFLOW
  - Run pm connectors inspect justcall before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
