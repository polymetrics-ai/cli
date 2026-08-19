---
name: pm-ticketmaster
description: Ticketmaster connector knowledge and safe action guide.
---

# pm-ticketmaster

## Purpose

Reads events, venues, attractions, and classifications from the Ticketmaster Discovery API.

## Icon

- id: ticketmaster
- asset: icons/ticketmaster.svg
- source: official
- review_status: official_verified
- review_url: https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- country_code
- keyword
- locale
- api_key (secret) (required)

## ETL Streams

- events:
  - primary key: id
  - fields: id(string), locale(string), name(string), type(string), url(string)
- venues:
  - primary key: id
  - fields: city(object), country(object), id(string), name(string), url(string)
- attractions:
  - primary key: id
  - fields: id(string), locale(string), name(string), type(string), url(string)
- classifications:
  - primary key: id
  - fields: genre(object), id(string), name(string), segment(object), subGenre(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Ticketmaster Discovery API read of public event/venue data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ticketmaster
```

### Inspect as structured JSON

```bash
pm connectors inspect ticketmaster --json
```

## Agent Rules

- Run pm connectors inspect ticketmaster before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
