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
  account_id
  base_url
  end_date
  mode
  start_date
  api_token (secret) (required)

ETL STREAMS
  accounts:
    primary key: id
    fields: currency(string), id(string), industry(string), name(string), status(string), time_zone(string), type(string)
  leads:
    primary key: id
    cursor: last_visit_date
    fields: city(string), country(string), employee_count(string), first_visit_date(string), id(string), industry(string), last_visit_date(string), name(string), quality(integer), type(string), visits(integer), website(string)
  visits:
    primary key: id
    cursor: visit_date
    fields: ended_at(string), hostname(string), id(string), pageviews(integer), referring_url(string), source(string), started_at(string), type(string), visit_date(string), visit_length(integer)
  custom_feeds:
    primary key: id
    fields: id(string), name(string), type(string)

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
