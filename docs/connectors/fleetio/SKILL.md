---
name: pm-fleetio
description: Fleetio connector knowledge and safe action guide.
---

# pm-fleetio

## Purpose

Reads Fleetio fleet management data: vehicles, contacts, fuel entries, issues, and service entries through the Fleetio REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_size
- account_token (secret)
- api_key (secret)

## ETL Streams

- vehicles:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(), created_at(), current_meter_value(), id(), license_plate(), make(), model(), name(), updated_at(), vehicle_status_name(), vehicle_type_name(), vin(), year()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(), created_at(), email(), employee(), first_name(), group_name(), id(), last_name(), name(), technician(), updated_at()
- fuel_entries:
  - primary key: id
  - cursor: updated_at
  - fields: cost(), created_at(), date(), id(), is_sample(), meter_value(), total_amount(), updated_at(), us_gallons(), vehicle_id(), vehicle_name()
- issues:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), due_date(), id(), number(), resolved_at(), state(), summary(), updated_at(), vehicle_id(), vehicle_name()
- service_entries:
  - primary key: id
  - cursor: updated_at
  - fields: completed_at(), created_at(), id(), is_sample(), labor_subtotal(), meter_value(), parts_subtotal(), started_at(), total_amount(), updated_at(), vehicle_id(), vehicle_name()

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
