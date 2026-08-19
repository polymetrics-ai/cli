---
name: pm-openweather
description: OpenWeather connector knowledge and safe action guide.
---

# pm-openweather

## Purpose

Reads current weather, hourly and daily forecasts, and government alerts for a configured geographic location from the OpenWeather One Call API 3.0.

## Icon

- id: openweather
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
- lat (required)
- lon (required)
- mode
- units
- appid (secret) (required)

## ETL Streams

- current:
  - primary key: lat, lon, dt
  - cursor: dt
  - fields: clouds(integer), dew_point(number), dt(integer), feels_like(number), humidity(integer), lat(string), lon(string), pressure(integer), sunrise(integer), sunset(integer), temp(number), timezone(string), uvi(number), visibility(integer), weather(array), wind_deg(integer), wind_gust(number), wind_speed(number)
- hourly:
  - primary key: lat, lon, dt
  - cursor: dt
  - fields: clouds(integer), dew_point(number), dt(integer), feels_like(number), humidity(integer), lat(string), lon(string), pop(number), pressure(integer), temp(number), timezone(string), uvi(number), visibility(integer), weather(array), wind_deg(integer), wind_gust(number), wind_speed(number)
- daily:
  - primary key: lat, lon, dt
  - cursor: dt
  - fields: dt(integer), humidity(integer), lat(string), lon(string), pop(number), pressure(integer), summary(string), sunrise(integer), sunset(integer), temp_day(number), temp_max(number), temp_min(number), timezone(string), uvi(number), weather(array), wind_deg(integer), wind_speed(number)
- alerts:
  - primary key: lat, lon, start, event
  - cursor: start
  - fields: description(string), end(integer), event(string), lat(string), lon(string), sender_name(string), start(integer), tags(array), timezone(string)

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
