---
name: pm-lucid-eld
description: Lucid ELD connector knowledge and safe action guide.
---

# pm-lucid-eld

## Purpose

Reads DriveHOS/Lucid ELD Partner API v2 company, driver, vehicle, latest status, and vehicle location history data through bounded read-only GET endpoints. The official API surface documents no mutations, reports, webhooks, or binary/media operations.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- end_date
- start_date
- vehicle_id
- company_api_key (secret)
- provider_api_key (secret)

## ETL Streams

- drivers:
- vehicles:
- vehicle_location_history:

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: read-only DriveHOS/Lucid ELD Partner API GET requests can return company, driver, vehicle, latest status, and vehicle location history data; outputs are bounded and secret-shaped fields are redacted by generic policies
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read DriveHOS/Lucid ELD Partner API v2 company, driver, vehicle, status, and vehicle-location data.
- Usage: pm lucid-eld <command> [flags]
- Source CLI: DriveHOS/Lucid ELD Partner API v2 (OpenAPI 2.0 fetched 2026-07-30)
- Global flags:
  - --credential (string): Credential name to use for the Lucid ELD request.
  - --connection (string): Alias for --credential.
  - --config (string_array): Connector config override as key=value, including vehicle_id/start_date/end_date for vehicle location history.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum ETL records to emit for stream commands.
  - --max-bytes (integer): Maximum direct-read response bytes; typed operations declare their own lower cap.
- Company
  - company info get - Retrieve Lucid ELD company info. [intent=direct_read availability=implemented]; approval: none: read-only bounded GET; risk: bounded read-only DriveHOS/Lucid ELD JSON read; secret-shaped fields are redacted.
- Drivers
  - drivers list - List Lucid ELD drivers as passthrough ETL records. [intent=etl availability=implemented stream=drivers]; notes: The stream schema is passthrough because the official OpenAPI response envelope leaves data untyped.; flags: --limit, --page
  - drivers get - Retrieve one Lucid ELD driver by driver_id. [intent=direct_read availability=implemented]; approval: none: read-only bounded GET; risk: bounded read-only DriveHOS/Lucid ELD JSON read; secret-shaped fields are redacted.; flags: --driver-id
- Vehicles
  - vehicles list - List Lucid ELD vehicles as passthrough ETL records. [intent=etl availability=implemented stream=vehicles]; notes: The stream schema is passthrough because the official OpenAPI response envelope leaves data untyped.; flags: --status, --limit, --page
  - vehicles get - Retrieve one Lucid ELD vehicle by vehicle_id. [intent=direct_read availability=implemented]; approval: none: read-only bounded GET; risk: bounded read-only DriveHOS/Lucid ELD JSON read; secret-shaped fields are redacted.; flags: --vehicle-id
- Vehicle location history
  - vehicle-location-history list - List vehicle location history for the configured vehicle_id and MM-DD-YYYY date window. [intent=etl availability=implemented stream=vehicle_location_history]; notes: vehicle_id is a connector config value for this stream path; start-date/end-date flags override the configured date window for the request query.; flags: --start-date, --end-date, --limit
- Latest status
  - latest driver statuses list - Retrieve latest Lucid ELD driver statuses with optional driver_id, limit, and page filters. [intent=direct_read availability=implemented]; approval: none: read-only bounded GET; risk: bounded read-only DriveHOS/Lucid ELD JSON read; secret-shaped fields are redacted.; flags: --driver-id, --limit, --page
  - latest vehicle statuses list - Retrieve latest Lucid ELD vehicle statuses with optional vehicle_id, limit, and page filters. [intent=direct_read availability=implemented]; approval: none: read-only bounded GET; risk: bounded read-only DriveHOS/Lucid ELD JSON read; secret-shaped fields are redacted.; flags: --vehicle-id, --limit, --page
- Help topics:
  - lucid-eld - Lucid ELD connector commands are definition-owned and read-only; no official mutation endpoints are documented.

## Commands

### Inspect as a manual

```bash
pm connectors inspect lucid-eld
```

### Inspect as structured JSON

```bash
pm connectors inspect lucid-eld --json
```

## Agent Rules

- Run pm connectors inspect lucid-eld before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
