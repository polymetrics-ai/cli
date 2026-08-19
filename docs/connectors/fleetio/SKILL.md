---
name: pm-fleetio
description: Fleetio connector knowledge and safe action guide.
---

# pm-fleetio

## Purpose

Reads Fleetio fleet management data: vehicles, contacts, fuel entries, issues, and service entries through the Fleetio REST API.

## Icon

- id: pm-sample
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
- page_size
- account_token (secret) (required)
- api_key (secret) (required)

## ETL Streams

- vehicles:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(string), created_at(string), current_meter_value(string), id(integer), license_plate(string), make(string), model(string), name(string), updated_at(string), vehicle_status_name(string), vehicle_type_name(string), vin(string), year(integer)
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(string), created_at(string), email(string), employee(boolean), first_name(string), group_name(string), id(integer), last_name(string), name(string), technician(boolean), updated_at(string)
- fuel_entries:
  - primary key: id
  - cursor: updated_at
  - fields: cost(string), created_at(string), date(string), id(integer), is_sample(boolean), meter_value(string), total_amount(string), updated_at(string), us_gallons(string), vehicle_id(integer), vehicle_name(string)
- issues:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), due_date(string), id(integer), number(integer), resolved_at(string), state(string), summary(string), updated_at(string), vehicle_id(integer), vehicle_name(string)
- service_entries:
  - primary key: id
  - cursor: updated_at
  - fields: completed_at(string), created_at(string), id(integer), is_sample(boolean), labor_subtotal(string), meter_value(string), parts_subtotal(string), started_at(string), total_amount(string), updated_at(string), vehicle_id(integer), vehicle_name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Fleetio API read of vehicle, contact, fuel entry, issue, and service entry data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect fleetio
```

### Inspect as structured JSON

```bash
pm connectors inspect fleetio --json
```

## Agent Rules

- Run pm connectors inspect fleetio before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
