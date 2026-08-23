---
name: pm-openaq
description: OpenAQ connector knowledge and safe action guide.
---

# pm-openaq

## Purpose

Reads OpenAQ air quality reference data (countries, parameters, locations, instruments, and manufacturers) from the OpenAQ v3 REST API.

## Icon

- id: openaq
- asset: icons/openaq.svg
- source: official
- review_status: official_verified
- review_url: https://docs.openaq.org/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- countries_id
- mode
- api_key (secret) (required)

## ETL Streams

- countries:
  - primary key: id
  - fields: code(string), datetimeFirst(string), datetimeLast(string), id(integer), name(string), parameters(array)
- parameters:
  - primary key: id
  - fields: description(string), displayName(string), id(integer), name(string), units(string)
- locations:
  - primary key: id
  - fields: coordinates(object), country(object), datetimeFirst(object), datetimeLast(object), id(integer), isMobile(boolean), isMonitor(boolean), locality(string), name(string), owner(object), provider(object), sensors(array), timezone(string)
- instruments:
  - primary key: id
  - fields: id(integer), isMonitor(boolean), manufacturer(object), name(string)
- manufacturers:
  - primary key: id
  - fields: id(integer), instruments(array), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external OpenAQ API read of public air-quality reference data
- approval: none; read-only public reference API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect openaq
```

### Inspect as structured JSON

```bash
pm connectors inspect openaq --json
```

## Agent Rules

- Run pm connectors inspect openaq before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
