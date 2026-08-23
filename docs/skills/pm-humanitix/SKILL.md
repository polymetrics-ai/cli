---
name: pm-humanitix
description: Humanitix connector knowledge and safe action guide.
---

# pm-humanitix

## Purpose

Reads Humanitix events, orders, tickets, and tags through the Humanitix public REST API.

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
- event_id
- page_size
- since
- api_key (secret) (required)

## ETL Streams

- events:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(string), createdAt(string), currency(string), endDate(string), location(string), markedAsSoldOut(boolean), name(string), organiserId(string), public(boolean), published(boolean), slug(string), startDate(string), updatedAt(string), userId(string)
- tags:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(string), createdAt(string), location(string), name(string), updatedAt(string), userId(string)
- orders:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(string), completedAt(string), createdAt(string), currency(string), email(string), eventDateId(string), eventId(string), financialStatus(string), firstName(string), lastName(string), manualOrder(boolean), mobile(string), orderName(string), status(string), total(number), updatedAt(string)
- tickets:
  - primary key: _id
  - cursor: updatedAt
  - fields: _id(string), createdAt(string), currency(string), eventDateId(string), eventId(string), firstName(string), isDonation(boolean), lastName(string), number(number), orderId(string), orderName(string), price(number), status(string), ticketTypeId(string), ticketTypeName(string), total(number), updatedAt(string)

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
