# pm connectors inspect freshbooks

```text
NAME
  pm connectors inspect freshbooks - FreshBooks connector manual

SYNOPSIS
  pm connectors inspect freshbooks
  pm connectors inspect freshbooks --json
  pm credentials add <name> --connector freshbooks [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads FreshBooks clients, invoices, expenses, payments, and items through the FreshBooks accounting REST API.

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
  account_id (required)
  base_url
  max_pages
  mode
  page_size
  oauth_access_token (secret) (required)

ETL STREAMS
  clients:
    primary key: id
    cursor: updated
    fields: currency_code(string), email(string), fname(string), id(integer), lname(string), organization(string), updated(string), userid(integer), vis_state(integer)
  invoices:
    primary key: id
    cursor: updated
    fields: amount(object), create_date(string), currency_code(string), customerid(integer), id(integer), invoice_number(string), invoiceid(integer), outstanding(object), status(integer), updated(string)
  expenses:
    primary key: id
    cursor: updated
    fields: amount(object), categoryid(integer), date(string), expenseid(integer), id(integer), notes(string), staffid(integer), updated(string), vendor(string)
  payments:
    primary key: id
    cursor: updated
    fields: amount(object), date(string), id(integer), invoiceid(integer), note(string), type(string), updated(string)
  items:
    primary key: id
    cursor: updated
    fields: description(string), id(integer), inventory(string), itemid(integer), name(string), qty(string), unit_cost(object), updated(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external FreshBooks API read of accounting data (clients, invoices, expenses, payments, items)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect freshbooks

  # Inspect as structured JSON
  pm connectors inspect freshbooks --json

AGENT WORKFLOW
  - Run pm connectors inspect freshbooks before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
