# pm connectors inspect openweather

```text
NAME
  pm connectors inspect openweather - OpenWeather connector manual

SYNOPSIS
  pm connectors inspect openweather
  pm connectors inspect openweather --json
  pm credentials add <name> --connector openweather [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads current weather, hourly and daily forecasts, and government alerts for a configured geographic location from the OpenWeather One Call API 3.0.

ICON
  id: openweather
  asset: icons/openweather.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://openweathermap.org/api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  lang
  lat (required)
  lon (required)
  mode
  units
  appid (secret) (required)

ETL STREAMS
  current:
    primary key: lat, lon, dt
    cursor: dt
    fields: clouds(integer), dew_point(number), dt(integer), feels_like(number), humidity(integer), lat(string), lon(string), pressure(integer), sunrise(integer), sunset(integer), temp(number), timezone(string), uvi(number), visibility(integer), weather(array), wind_deg(integer), wind_gust(number), wind_speed(number)
  hourly:
    primary key: lat, lon, dt
    cursor: dt
    fields: clouds(integer), dew_point(number), dt(integer), feels_like(number), humidity(integer), lat(string), lon(string), pop(number), pressure(integer), temp(number), timezone(string), uvi(number), visibility(integer), weather(array), wind_deg(integer), wind_gust(number), wind_speed(number)
  daily:
    primary key: lat, lon, dt
    cursor: dt
    fields: dt(integer), humidity(integer), lat(string), lon(string), pop(number), pressure(integer), summary(string), sunrise(integer), sunset(integer), temp_day(number), temp_max(number), temp_min(number), timezone(string), uvi(number), weather(array), wind_deg(integer), wind_speed(number)
  alerts:
    primary key: lat, lon, start, event
    cursor: start
    fields: description(string), end(integer), event(string), lat(string), lon(string), sender_name(string), start(integer), tags(array), timezone(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external OpenWeather API read of public weather forecast data
  approval: none; read-only public weather API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect openweather

  # Inspect as structured JSON
  pm connectors inspect openweather --json

AGENT WORKFLOW
  - Run pm connectors inspect openweather before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
