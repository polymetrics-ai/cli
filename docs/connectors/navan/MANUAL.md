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
  base_url
  mode
  start_date
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(string), base_price(number), booking_fee(number), booking_id(string), booking_method(string), booking_status(string), booking_type(string), cancelled_at(string), confirmation_number(string), created(string), currency(string), destination(string), domestic(boolean), end_date(string), expensed(boolean), grand_total(number), last_modified(string), start_date(string), uuid(string)
  hotel_bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(string), base_price(number), booking_fee(number), booking_id(string), booking_method(string), booking_status(string), booking_type(string), cancelled_at(string), confirmation_number(string), created(string), currency(string), destination(string), domestic(boolean), end_date(string), expensed(boolean), grand_total(number), last_modified(string), start_date(string), uuid(string)
  car_bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(string), base_price(number), booking_fee(number), booking_id(string), booking_method(string), booking_status(string), booking_type(string), cancelled_at(string), confirmation_number(string), created(string), currency(string), destination(string), domestic(boolean), end_date(string), expensed(boolean), grand_total(number), last_modified(string), start_date(string), uuid(string)
  rail_bookings:
    primary key: uuid
    cursor: last_modified
    fields: approval_status(string), base_price(number), booking_fee(number), booking_id(string), booking_method(string), booking_status(string), booking_type(string), cancelled_at(string), confirmation_number(string), created(string), currency(string), destination(string), domestic(boolean), end_date(string), expensed(boolean), grand_total(number), last_modified(string), start_date(string), uuid(string)

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
