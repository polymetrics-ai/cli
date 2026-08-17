---
name: pm-hubplanner
description: Hubplanner connector knowledge and safe action guide.
---

# pm-hubplanner

## Purpose

Reads Hubplanner resources, projects, clients, events, holidays, bookings, and billing rates through the Hubplanner REST API.

## Icon

- asset: icons/hubplanner.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- api_key (secret)

## ETL Streams

- resources:
  - primary key: _id
  - fields: _id(), createdDate(), email(), firstName(), lastName(), note(), role(), status(), type()
- projects:
  - primary key: _id
  - fields: _id(), budgetCashAmount(), budgetCurrency(), budgetHours(), createdDate(), name(), note(), projectCode(), status(), updatedDate()
- clients:
  - primary key: _id
  - fields: _id(), createdDate(), email(), name(), note(), phone()
- events:
  - primary key: _id
  - fields: _id(), end(), name(), note(), start(), type()
- holidays:
  - primary key: _id
  - fields: _id(), date(), end(), holidayGroup(), name(), start()
- bookings:
  - primary key: _id
  - fields: _id(), category(), end(), note(), project(), resource(), start(), state()
- billing_rates:
  - primary key: _id
  - fields: _id(), currency(), default(), name(), rate()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Hubplanner API read of scheduling, project, and billing data
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect hubplanner
```

### Inspect as structured JSON

```bash
pm connectors inspect hubplanner --json
```

## Agent Rules

- Run pm connectors inspect hubplanner before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
