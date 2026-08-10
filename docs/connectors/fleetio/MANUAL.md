# pm connectors inspect fleetio

```text
NAME
  pm connectors inspect fleetio - Fleetio connector manual

SYNOPSIS
  pm connectors inspect fleetio
  pm connectors inspect fleetio --json
  pm credentials add <name> --connector fleetio [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Fleetio fleet management data: vehicles, contacts, fuel entries, issues, and service entries through the Fleetio REST API.

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
  page_size
  account_token (secret) (required)
  api_key (secret) (required)

ETL STREAMS
  vehicles:
    primary key: id
    cursor: updated_at
    fields: archived_at(string), created_at(string), current_meter_value(string), id(integer), license_plate(string), make(string), model(string), name(string), updated_at(string), vehicle_status_name(string), vehicle_type_name(string), vin(string), year(integer)
  contacts:
    primary key: id
    cursor: updated_at
    fields: archived_at(string), created_at(string), email(string), employee(boolean), first_name(string), group_name(string), id(integer), last_name(string), name(string), technician(boolean), updated_at(string)
  fuel_entries:
    primary key: id
    cursor: updated_at
    fields: cost(string), created_at(string), date(string), id(integer), is_sample(boolean), meter_value(string), total_amount(string), updated_at(string), us_gallons(string), vehicle_id(integer), vehicle_name(string)
  issues:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), due_date(string), id(integer), number(integer), resolved_at(string), state(string), summary(string), updated_at(string), vehicle_id(integer), vehicle_name(string)
  service_entries:
    primary key: id
    cursor: updated_at
    fields: completed_at(string), created_at(string), id(integer), is_sample(boolean), labor_subtotal(string), meter_value(string), parts_subtotal(string), started_at(string), total_amount(string), updated_at(string), vehicle_id(integer), vehicle_name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Fleetio API read of vehicle, contact, fuel entry, issue, and service entry data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect fleetio

  # Inspect as structured JSON
  pm connectors inspect fleetio --json

AGENT WORKFLOW
  - Run pm connectors inspect fleetio before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
