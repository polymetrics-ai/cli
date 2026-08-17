# pm connectors inspect navan

```text
NAME
  pm connectors inspect navan - Navan connector manual

SYNOPSIS
  pm connectors inspect navan
  pm connectors inspect navan --json
  pm credentials add <name> --connector navan [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Navan flight, hotel, car, and rail travel bookings through the Navan REST API using OAuth2 client-credentials authentication.

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
  mode
  start_date
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()
  hotel_bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()
  car_bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()
  rail_bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Navan API read of travel booking data (flight, hotel, car, rail)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect navan

  # Inspect as structured JSON
  pm connectors inspect navan --json

AGENT WORKFLOW
  - Run pm connectors inspect navan before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
