# pm connectors inspect breezometer

```text
NAME
  pm connectors inspect breezometer - Breezometer connector manual

SYNOPSIS
  pm connectors inspect breezometer
  pm connectors inspect breezometer --json
  pm credentials add <name> --connector breezometer [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads BreezoMeter (Google Environment) air quality, pollen, weather, and wildfire-tracking conditions/forecasts for a configured location via the BreezoMeter REST API; writes a stateless cleanest-route environmental-cleanliness scoring computation.

ICON
  asset: icons/breezometer.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.breezometer.com/api-documentation/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  days_to_forecast
  historic_hours
  hours_to_forecast
  latitude
  longitude
  mode
  weather_days_to_forecast
  wildfire_days_from_extinguish
  wildfire_radius_km
  api_key (secret)

ETL STREAMS
  air_quality_current:
    primary key: datetime, latitude, longitude
    cursor: datetime
    fields: data_available(), datetime(), health_recommendations(), indexes(), latitude(), longitude(), pollutants()
  air_quality_forecast:
    primary key: datetime, latitude, longitude
    cursor: datetime
    fields: data_available(), datetime(), health_recommendations(), indexes(), latitude(), longitude(), pollutants()
  air_quality_history:
    primary key: datetime, latitude, longitude
    cursor: datetime
    fields: data_available(), datetime(), health_recommendations(), indexes(), latitude(), longitude(), pollutants()
  pollen_forecast:
    primary key: datetime, latitude, longitude
    cursor: datetime
    fields: data_available(), date(), datetime(), index(), latitude(), longitude(), plants(), types()
  weather_current:
    primary key: datetime, latitude, longitude
    cursor: datetime
    fields: data_available(), datetime(), feels_like_temperature(), latitude(), longitude(), precipitation(), relative_humidity(), temperature(), weather_condition(), wind()
  weather_daily_forecast:
    primary key: start_date, latitude, longitude
    cursor: start_date
    fields: latitude(), longitude(), max_uv_index(), moon(), start_date(), sun()
  wildfire_active_tracking:
    primary key: EventId, latitude, longitude
    cursor: LastUpdated
    fields: CalculatedAcres(), CurrentLat(), CurrentLon(), DiscoveryDateTime(), EventId(), ExistenceConfidence(), LastUpdated(), MaxCalculatedAcres(), ShapeConfidence(), geometry(), latitude(), longitude()
  wildfire_burnt_area:
    primary key: latitude, longitude, DiscoveryDateTime
    cursor: ExtinguishedTS
    fields: BurntAcres(), BurntLat(), BurntLon(), DiscoveryDateTime(), ExtinguishedTS(), InitialLat(), InitialLon(), geometry(), latitude(), longitude()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  score_cleanest_route:
    endpoint: POST /insights/v1/cleanest-route
    risk: stateless environmental-cleanliness scoring computation over caller-supplied route geometries; creates or mutates no persistent BreezoMeter object and has no side effects beyond the API call itself, low-risk

SECURITY
  read risk: external BreezoMeter API read of point-in-time environmental, weather, and wildfire-tracking data for the configured location
  write risk: stateless environmental-cleanliness scoring computation over caller-supplied route geometries; no persistent BreezoMeter object is created or mutated
  approval: none; the sole write action is a stateless computation with no side effects on BreezoMeter's own data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect breezometer

  # Inspect as structured JSON
  pm connectors inspect breezometer --json

AGENT WORKFLOW
  - Run pm connectors inspect breezometer before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
