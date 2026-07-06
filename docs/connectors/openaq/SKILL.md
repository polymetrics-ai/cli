---
name: pm-openaq
description: OpenAQ connector knowledge and safe action guide.
---

# pm-openaq

## Purpose

Reads OpenAQ air quality reference data (countries, parameters, locations, instruments, and manufacturers) from the OpenAQ v3 REST API.

## Icon

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
- api_key (secret)

## ETL Streams

- countries:
  - primary key: id
  - fields: code(), datetimeFirst(), datetimeLast(), id(), name(), parameters()
- parameters:
  - primary key: id
  - fields: description(), displayName(), id(), name(), units()
- locations:
  - primary key: id
  - fields: coordinates(), country(), datetimeFirst(), datetimeLast(), id(), isMobile(), isMonitor(), locality(), name(), owner(), provider(), sensors(), timezone()
- instruments:
  - primary key: id
  - fields: id(), isMonitor(), manufacturer(), name()
- manufacturers:
  - primary key: id
  - fields: id(), instruments(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
