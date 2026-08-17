# pm connectors inspect acuity-scheduling

```text
NAME
  pm connectors inspect acuity-scheduling - Acuity Scheduling connector manual

SYNOPSIS
  pm connectors inspect acuity-scheduling
  pm connectors inspect acuity-scheduling --json
  pm credentials add <name> --connector acuity-scheduling [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Acuity Scheduling appointments, clients, appointment types, calendars, forms, products, orders, and labels, and writes appointment/block/certificate mutations, through the Acuity REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  username
  password (secret)

ETL STREAMS
  appointments:
    primary key: id
    cursor: datetime
    fields: amount_paid(), appointment_type_id(), calendar(), calendar_id(), canceled(), date(), datetime(), datetime_created(), duration(), email(), end_time(), first_name(), id(), last_name(), paid(), phone(), price(), time(), type()
  clients:
    primary key: email
    fields: email(), first_name(), last_name(), phone()
  appointment_types:
    primary key: id
    fields: active(), category(), color(), description(), duration(), id(), name(), price(), private(), type()
  calendars:
    primary key: id
    fields: description(), email(), id(), location(), name(), replyTo(), timezone()
  forms:
    primary key: id
    fields: description(), hidden(), id(), name()
  products:
    primary key: id
    fields: description(), expires(), hidden(), id(), minutes(), name(), price(), type()
  orders:
    primary key: id
    fields: email(), first_name(), id(), last_name(), notes(), phone(), status(), time(), title(), total()
  labels:
    primary key: id
    fields: color(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_appointment:
    endpoint: POST /appointments
    risk: creates a live appointment booking on the calendar and, depending on account settings, sends the client a confirmation email/SMS; external mutation, approval required
  update_appointment:
    endpoint: PUT /appointments/{{ record.id }}
    required fields: id
    risk: updates a live appointment's client-facing details from Acuity's white-list of updatable attributes; external mutation, approval required
  cancel_appointment:
    endpoint: PUT /appointments/{{ record.id }}/cancel
    required fields: id
    risk: permanently cancels a live scheduled appointment; irreversible (Acuity's own docs: it is not possible to un-cancel), and by default sends the client a cancellation notification. External mutation, approval required
  create_block:
    endpoint: POST /blocks
    risk: blocks off a time range on a live calendar, preventing clients from booking appointments in it; external mutation, approval required
  create_certificate:
    endpoint: POST /certificates
    risk: issues a live, redeemable package or coupon certificate code; external mutation, approval required

SECURITY
  read risk: external Acuity Scheduling API read of appointments, clients, appointment types, calendars, forms, products, orders, and labels
  write risk: external Acuity Scheduling mutation: creates/updates/cancels live appointments, blocks off calendar time, and issues package/coupon certificates; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect acuity-scheduling

  # Inspect as structured JSON
  pm connectors inspect acuity-scheduling --json

AGENT WORKFLOW
  - Run pm connectors inspect acuity-scheduling before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
