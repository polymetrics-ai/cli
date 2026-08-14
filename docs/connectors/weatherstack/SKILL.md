---
name: pm-weatherstack
description: Weatherstack connector knowledge and safe action guide.
---

# pm-weatherstack

## Purpose

Reads current, historical, forecast, marine, and location-autocomplete weather data from Weatherstack. Read-only.

## Icon

- id: weatherstack
- asset: icons/weatherstack.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://weatherstack.com/documentation

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- autocomplete_query
- base_url
- forecast_days
- historical_date
- language
- latitude
- longitude
- mode
- query (required)
- units
- access_key (secret) (required)

## ETL Streams

- current:
  - primary key: id
  - fields: current(object), id(string), location(object)
- historical:
  - primary key: id
  - fields: historical(object), id(string), location(object)
- forecast:
  - primary key: id
  - fields: forecast(object), id(string), location(object)
- marine:
  - primary key: id
  - fields: current(object), id(string), location(object)
- autocomplete:
  - primary key: name, region, country, lat, lon
  - fields: country(string), lat(string), lon(string), name(string), region(string), timezone_id(string), utc_offset(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Weatherstack API read of public weather data
- approval: none; read-only public weather data connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect weatherstack
```

### Inspect as structured JSON

```bash
pm connectors inspect weatherstack --json
```

## Agent Rules

- Run pm connectors inspect weatherstack before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
