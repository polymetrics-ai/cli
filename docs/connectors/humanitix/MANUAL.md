# pm connectors inspect humanitix

```text
NAME
  pm connectors inspect humanitix - Humanitix connector manual

SYNOPSIS
  pm connectors inspect humanitix
  pm connectors inspect humanitix --json
  pm credentials add <name> --connector humanitix [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Humanitix events, orders, tickets, and tags through the Humanitix public REST API.

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
  page_size
  since
  api_key (secret)

ETL STREAMS
  events:
    primary key: _id
    cursor: updatedAt
    fields: _id(), createdAt(), currency(), endDate(), location(), markedAsSoldOut(), name(), organiserId(), public(), published(), slug(), startDate(), updatedAt(), userId()
  tags:
    primary key: _id
    cursor: updatedAt
    fields: _id(), createdAt(), location(), name(), updatedAt(), userId()
  orders:
    primary key: _id
    cursor: updatedAt
    fields: _id(), completedAt(), createdAt(), currency(), email(), eventDateId(), eventId(), financialStatus(), firstName(), lastName(), manualOrder(), mobile(), orderName(), status(), total(), updatedAt()
  tickets:
    primary key: _id
    cursor: updatedAt
    fields: _id(), createdAt(), currency(), eventDateId(), eventId(), firstName(), isDonation(), lastName(), number(), orderId(), orderName(), price(), status(), ticketTypeId(), ticketTypeName(), total(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Humanitix API read of event, order, ticket, and tag data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect humanitix

  # Inspect as structured JSON
  pm connectors inspect humanitix --json

AGENT WORKFLOW
  - Run pm connectors inspect humanitix before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
