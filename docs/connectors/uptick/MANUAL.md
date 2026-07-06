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
  page_size
  start_date
  username
  client_id (secret)
  client_secret (secret)
  password (secret)

ETL STREAMS
  tasks:
    primary key: id
    cursor: updated
    fields: client(), created(), deleted(), description(), due(), id(), is_active(), name(), priority(), property(), ref(), status(), updated()
  clients:
    primary key: id
    cursor: updated
    fields: address(), contact_email(), contact_name(), contact_phone_bh(), created(), id(), is_active(), name(), notes(), ref(), updated()
  properties:
    primary key: id
    cursor: updated
    fields: address(), coords(), created(), id(), name(), ref(), status(), timezone(), updated()
  invoices:
    primary key: id
    cursor: updated
    fields: created(), currency(), date(), description(), due_date(), gst(), id(), is_overdue(), is_sent(), number(), property(), ref(), status(), subtotal(), task(), total(), updated()
  assets:
    primary key: id
    cursor: updated
    fields: barcode(), created(), deleted(), id(), is_active(), label(), location(), make(), model(), property(), ref(), serviced_date(), size(), status(), type(), updated(), uptick_ref(), variant()
  quotes:
    primary key: id
    fields: created(), description(), id(), ref(), status(), total(), updated()
  purchaseorders:
    primary key: id
    fields: created(), id(), ref(), status(), supplier(), total(), updated()
  forms:
    primary key: id
    fields: created(), description(), id(), name(), status(), updated()
  users:
    primary key: id
    fields: created(), email(), id(), is_active(), name(), updated(), username()
  teams:
    primary key: id
    fields: created(), description(), id(), is_active(), name(), updated()
  stockitems:
    primary key: id
    fields: created(), description(), id(), is_active(), name(), ref(), updated()

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
