# pm connectors inspect oncehub

```text
NAME
  pm connectors inspect oncehub - OnceHub connector manual

SYNOPSIS
  pm connectors inspect oncehub
  pm connectors inspect oncehub --json
  pm credentials add <name> --connector oncehub [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads OnceHub bookings, contacts, booking pages, users, and event types through the OnceHub REST API.

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
  max_pages
  mode
  page_size
  start_date
  api_key (secret)

ETL STREAMS
  bookings:
    primary key: id
    cursor: last_updated_time
    fields: booking_page(), contact(), creation_time(), customer_timezone(), duration_minutes(), event_type(), id(), in_trash(), last_updated_time(), location_description(), object(), owner(), starting_time(), status(), subject(), tracking_id()
  contacts:
    primary key: id
    cursor: last_updated_time
    fields: creation_time(), email(), first_name(), id(), last_updated_time(), mobile_phone(), object(), owner(), timezone()
  booking_pages:
    primary key: id
    fields: active(), id(), label(), name(), object(), timezone(), url()
  users:
    primary key: id
    fields: email(), first_name(), id(), last_name(), object(), role_name(), status()
  event_types:
    primary key: id
    fields: id(), label(), name(), object()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external OnceHub API read of scheduling, contact, and user data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect oncehub

  # Inspect as structured JSON
  pm connectors inspect oncehub --json

AGENT WORKFLOW
  - Run pm connectors inspect oncehub before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
