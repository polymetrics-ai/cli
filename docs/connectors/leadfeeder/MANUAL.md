# pm connectors inspect leadfeeder

```text
NAME
  pm connectors inspect leadfeeder - Leadfeeder connector manual

SYNOPSIS
  pm connectors inspect leadfeeder
  pm connectors inspect leadfeeder --json
  pm credentials add <name> --connector leadfeeder [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Leadfeeder accounts and their leads, visits, and custom feeds through the Leadfeeder JSON:API.

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
  account_id
  base_url
  end_date
  mode
  start_date
  api_token (secret)

ETL STREAMS
  accounts:
    primary key: id
    fields: currency(), id(), industry(), name(), status(), time_zone(), type()
  leads:
    primary key: id
    cursor: last_visit_date
    fields: city(), country(), employee_count(), first_visit_date(), id(), industry(), last_visit_date(), name(), quality(), type(), visits(), website()
  visits:
    primary key: id
    cursor: visit_date
    fields: ended_at(), hostname(), id(), pageviews(), referring_url(), source(), started_at(), type(), visit_date(), visit_length()
  custom_feeds:
    primary key: id
    fields: id(), name(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Leadfeeder API read of account, lead, and visit data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect leadfeeder

  # Inspect as structured JSON
  pm connectors inspect leadfeeder --json

AGENT WORKFLOW
  - Run pm connectors inspect leadfeeder before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
