---
name: pm-openweather
description: OpenWeather connector knowledge and safe action guide.
---

# pm-openweather

## Purpose

Reads current weather, hourly and daily forecasts, and government alerts for a configured geographic location from the OpenWeather One Call API 3.0.

## Icon

- asset: icons/openweather.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://openweathermap.org/api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- lang
- lat
- lon
- mode
- units
- appid (secret)

## ETL Streams

- current:
  - primary key: lat, lon, dt
  - cursor: dt
  - fields: clouds(), dew_point(), dt(), feels_like(), humidity(), lat(), lon(), pressure(), sunrise(), sunset(), temp(), timezone(), uvi(), visibility(), weather(), wind_deg(), wind_gust(), wind_speed()
- hourly:
  - primary key: lat, lon, dt
  - cursor: dt
  - fields: clouds(), dew_point(), dt(), feels_like(), humidity(), lat(), lon(), pop(), pressure(), temp(), timezone(), uvi(), visibility(), weather(), wind_deg(), wind_gust(), wind_speed()
- daily:
  - primary key: lat, lon, dt
  - cursor: dt
  - fields: dt(), humidity(), lat(), lon(), pop(), pressure(), summary(), sunrise(), sunset(), temp_day(), temp_max(), temp_min(), timezone(), uvi(), weather(), wind_deg(), wind_speed()
- alerts:
  - primary key: lat, lon, start, event
  - cursor: start
  - fields: description(), end(), event(), lat(), lon(), sender_name(), start(), tags(), timezone()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external OpenWeather API read of public weather forecast data
- approval: none; read-only public weather API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect openweather
```

### Inspect as structured JSON

```bash
pm connectors inspect openweather --json
```

## Agent Rules

- Run pm connectors inspect openweather before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
