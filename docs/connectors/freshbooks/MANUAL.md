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
  max_pages
  mode
  page_size
  oauth_access_token (secret)

ETL STREAMS
  clients:
    primary key: id
    cursor: updated
    fields: currency_code(), email(), fname(), id(), lname(), organization(), updated(), userid(), vis_state()
  invoices:
    primary key: id
    cursor: updated
    fields: amount(), create_date(), currency_code(), customerid(), id(), invoice_number(), invoiceid(), outstanding(), status(), updated()
  expenses:
    primary key: id
    cursor: updated
    fields: amount(), categoryid(), date(), expenseid(), id(), notes(), staffid(), updated(), vendor()
  payments:
    primary key: id
    cursor: updated
    fields: amount(), date(), id(), invoiceid(), note(), type(), updated()
  items:
    primary key: id
    cursor: updated
    fields: description(), id(), inventory(), itemid(), name(), qty(), unit_cost(), updated()

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
