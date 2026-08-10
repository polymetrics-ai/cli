# pm connectors inspect repairshopr

```text
NAME
  pm connectors inspect repairshopr - RepairShopr connector manual

SYNOPSIS
  pm connectors inspect repairshopr
  pm connectors inspect repairshopr --json
  pm credentials add <name> --connector repairshopr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads RepairShopr customers, tickets, invoices, estimates, and assets through the REST API.

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
  base_url (required)
  created_after
  query
  updated_after
  api_token (secret) (required)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), id(string), name(string), phone(string), stream(string), updated_at(string)
  tickets:
    primary key: id
    cursor: updated_at
    fields: customer_id(string), id(string), number(string), status(string), stream(string), subject(string), updated_at(string)
  invoices:
    primary key: id
    cursor: updated_at
    fields: customer_id(string), id(string), number(string), status(string), stream(string), total(string), updated_at(string)
  estimates:
    primary key: id
    cursor: updated_at
    fields: customer_id(string), id(string), number(string), status(string), stream(string), total(string), updated_at(string)
  assets:
    primary key: id
    cursor: updated_at
    fields: customer_id(string), id(string), name(string), serial_number(string), stream(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external RepairShopr API read of customer and shop-management data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect repairshopr

  # Inspect as structured JSON
  pm connectors inspect repairshopr --json

AGENT WORKFLOW
  - Run pm connectors inspect repairshopr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
