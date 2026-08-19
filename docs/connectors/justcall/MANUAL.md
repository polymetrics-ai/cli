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
  mode
  page_size
  start_date
  api_key_2 (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: available(string), created_at(string), email(string), extension(string), id(string), last_login_timestamp(string), name(string), on_call(string), role(string), timezone(string)
  calls:
    primary key: id
    cursor: call_date
    fields: agent_email(string), agent_id(string), agent_name(string), call_date(string), call_duration(string), call_sid(string), call_time(string), contact_name(string), contact_number(string), cost_incurred(string), id(string), justcall_line_name(string), justcall_number(string)
  sms:
    primary key: id
    cursor: sms_date
    fields: agent_email(string), agent_id(string), agent_name(string), contact_name(string), contact_number(string), cost_incurred(string), delivery_status(string), direction(string), id(string), justcall_line_name(string), justcall_number(string), sms_date(string), sms_time(string)

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
