# pm connectors inspect hubplanner

```text
NAME
  pm connectors inspect hubplanner - Hubplanner connector manual

SYNOPSIS
  pm connectors inspect hubplanner
  pm connectors inspect hubplanner --json
  pm credentials add <name> --connector hubplanner [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Hubplanner resources, projects, clients, events, holidays, bookings, and billing rates through the Hubplanner REST API.

ICON
  asset: icons/hubplanner.svg
  source: upstream_registry
  review_status: upstream_seeded

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
  api_key (secret)

ETL STREAMS
  resources:
    primary key: _id
    fields: _id(), createdDate(), email(), firstName(), lastName(), note(), role(), status(), type()
  projects:
    primary key: _id
    fields: _id(), budgetCashAmount(), budgetCurrency(), budgetHours(), createdDate(), name(), note(), projectCode(), status(), updatedDate()
  clients:
    primary key: _id
    fields: _id(), createdDate(), email(), name(), note(), phone()
  events:
    primary key: _id
    fields: _id(), end(), name(), note(), start(), type()
  holidays:
    primary key: _id
    fields: _id(), date(), end(), holidayGroup(), name(), start()
  bookings:
    primary key: _id
    fields: _id(), category(), end(), note(), project(), resource(), start(), state()
  billing_rates:
    primary key: _id
    fields: _id(), currency(), default(), name(), rate()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Hubplanner API read of scheduling, project, and billing data
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect hubplanner

  # Inspect as structured JSON
  pm connectors inspect hubplanner --json

AGENT WORKFLOW
  - Run pm connectors inspect hubplanner before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
