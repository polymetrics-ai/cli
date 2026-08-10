# pm connectors inspect uptick

```text
NAME
  pm connectors inspect uptick - Uptick connector manual

SYNOPSIS
  pm connectors inspect uptick
  pm connectors inspect uptick --json
  pm credentials add <name> --connector uptick [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Uptick field service management data through the Uptick REST API using OAuth2 password-grant auth.

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
  page_size
  start_date
  username (required)
  client_id (secret) (required)
  client_secret (secret) (required)
  password (secret) (required)

ETL STREAMS
  tasks:
    primary key: id
    cursor: updated
    fields: client(string), created(string), deleted(string), description(string), due(string), id(integer), is_active(boolean), name(string), priority(string), property(string), ref(string), status(string), updated(string)
  clients:
    primary key: id
    cursor: updated
    fields: address(string), contact_email(string), contact_name(string), contact_phone_bh(string), created(string), id(integer), is_active(boolean), name(string), notes(string), ref(string), updated(string)
  properties:
    primary key: id
    cursor: updated
    fields: address(string), coords(string), created(string), id(integer), name(string), ref(string), status(string), timezone(string), updated(string)
  invoices:
    primary key: id
    cursor: updated
    fields: created(string), currency(string), date(string), description(string), due_date(string), gst(string), id(integer), is_overdue(boolean), is_sent(boolean), number(string), property(string), ref(string), status(string), subtotal(string), task(string), total(string), updated(string)
  assets:
    primary key: id
    cursor: updated
    fields: barcode(string), created(string), deleted(string), id(integer), is_active(boolean), label(string), location(string), make(string), model(string), property(string), ref(string), serviced_date(string), size(string), status(string), type(string), updated(string), uptick_ref(string), variant(string)
  quotes:
    primary key: id
    fields: created(string), description(string), id(integer), ref(string), status(string), total(integer), updated(string)
  purchaseorders:
    primary key: id
    fields: created(string), id(integer), ref(string), status(string), supplier(string), total(number), updated(string)
  forms:
    primary key: id
    fields: created(string), description(string), id(integer), name(string), status(string), updated(string)
  users:
    primary key: id
    fields: created(string), email(string), id(integer), is_active(boolean), name(string), updated(string), username(string)
  teams:
    primary key: id
    fields: created(string), description(string), id(integer), is_active(boolean), name(string), updated(string)
  stockitems:
    primary key: id
    fields: created(string), description(string), id(integer), is_active(boolean), name(string), ref(string), updated(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Uptick field service management API reads for tasks, clients, properties, invoices, assets, quotes, purchase orders, forms, users, teams, and stock items
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect uptick

  # Inspect as structured JSON
  pm connectors inspect uptick --json

AGENT WORKFLOW
  - Run pm connectors inspect uptick before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
