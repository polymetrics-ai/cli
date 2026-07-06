---
name: pm-humanitix
description: Humanitix connector knowledge and safe action guide.
---

# pm-humanitix

## Purpose

Reads Humanitix events, orders, tickets, and tags through the Humanitix public REST API.

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
- event_id
- page_size
- since
- api_key (secret)

## ETL Streams

- events:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(), createdAt(), currency(), endDate(), location(), markedAsSoldOut(), name(), organiserId(), public(), published(), slug(), startDate(), updatedAt(), userId()
- tags:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(), createdAt(), location(), name(), updatedAt(), userId()
- orders:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(), completedAt(), createdAt(), currency(), email(), eventDateId(), eventId(), financialStatus(), firstName(), lastName(), manualOrder(), mobile(), orderName(), status(), total(), updatedAt()
- tickets:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(), createdAt(), currency(), eventDateId(), eventId(), firstName(), isDonation(), lastName(), number(), orderId(), orderName(), price(), status(), ticketTypeId(), ticketTypeName(), total(), updatedAt()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Humanitix API read of event, order, ticket, and tag data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect humanitix
```

### Inspect as structured JSON

```bash
pm connectors inspect humanitix --json
```

## Agent Rules

- Run pm connectors inspect humanitix before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
