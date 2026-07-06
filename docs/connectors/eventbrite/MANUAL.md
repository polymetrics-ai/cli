# pm connectors inspect eventbrite

```text
NAME
  pm connectors inspect eventbrite - Eventbrite connector manual

SYNOPSIS
  pm connectors inspect eventbrite
  pm connectors inspect eventbrite --json
  pm credentials add <name> --connector eventbrite [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Eventbrite organizations, events, attendees, orders, and ticket classes through the Eventbrite v3 REST API. Read-only source.

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
  event_id
  organization_id
  start_date
  private_token (secret)

ETL STREAMS
  organizations:
    primary key: id
    fields: id(), image_id(), locale(), name(), vertical()
  events:
    primary key: id
    cursor: changed
    fields: capacity(), changed(), created(), currency(), description(), end(), id(), listed(), name(), online_event(), organization_id(), published(), start(), status(), url(), venue_id()
  attendees:
    primary key: id
    cursor: changed
    fields: cancelled(), changed(), checked_in(), created(), email(), event_id(), id(), name(), order_id(), quantity(), refunded(), status(), ticket_class_id(), ticket_class_name()
  orders:
    primary key: id
    cursor: changed
    fields: changed(), created(), email(), event_id(), id(), name(), status(), time_remaining()
  ticket_classes:
    primary key: id
    fields: cost(), description(), event_id(), fee(), free(), hidden(), id(), name(), on_sale_status(), quantity_sold(), quantity_total()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Eventbrite API read of organization, event, attendee, and order data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect eventbrite

  # Inspect as structured JSON
  pm connectors inspect eventbrite --json

AGENT WORKFLOW
  - Run pm connectors inspect eventbrite before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
