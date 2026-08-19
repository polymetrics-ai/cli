# pm connectors inspect jobnimbus

```text
NAME
  pm connectors inspect jobnimbus - JobNimbus connector manual

SYNOPSIS
  pm connectors inspect jobnimbus
  pm connectors inspect jobnimbus --json
  pm credentials add <name> --connector jobnimbus [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads JobNimbus CRM contacts, jobs, tasks, activities, and files through the JobNimbus REST API.

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
  api_key (secret) (required)

ETL STREAMS
  contacts:
    primary key: jnid
    cursor: date_updated
    fields: address_line1(string), city(string), company(string), country_name(string), date_created(integer), date_updated(integer), display_name(string), email(string), first_name(string), home_phone(string), is_active(boolean), is_archived(boolean), jnid(string), last_name(string), mobile_phone(string), record_type_name(string), sales_rep_name(string), state_text(string), status_name(string), type(string), work_phone(string), zip(string)
  jobs:
    primary key: jnid
    cursor: date_updated
    fields: address_line1(string), approved_estimate_total(number), approved_invoice_total(number), city(string), customer(string), date_created(integer), date_status_change(integer), date_updated(integer), is_active(boolean), is_archived(boolean), jnid(string), name(string), number(string), record_type_name(string), sales_rep_name(string), state_text(string), status_name(string), type(string), zip(string)
  tasks:
    primary key: jnid
    cursor: date_updated
    fields: customer(string), date_created(integer), date_end(integer), date_start(integer), date_updated(integer), description(string), is_active(boolean), is_archived(boolean), is_completed(boolean), jnid(string), number(string), priority(integer), record_type_name(string), title(string), type(string)
  activities:
    primary key: jnid
    cursor: date_updated
    fields: created_by_name(string), customer(string), date_created(integer), date_updated(integer), is_active(boolean), is_archived(boolean), is_status_change(boolean), jnid(string), note(string), record_type_name(string), sales_rep_name(string), source(string), type(string)
  files:
    primary key: jnid
    cursor: date_updated
    fields: content_type(string), created_by_name(string), customer(string), date_created(integer), date_file_created(integer), date_updated(integer), description(string), filename(string), is_active(boolean), jnid(string), md5(string), record_type_name(string), size(number), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external JobNimbus API read of CRM contact, job, task, activity, and file data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect jobnimbus

  # Inspect as structured JSON
  pm connectors inspect jobnimbus --json

AGENT WORKFLOW
  - Run pm connectors inspect jobnimbus before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
