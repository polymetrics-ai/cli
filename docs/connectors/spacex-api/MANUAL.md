# pm connectors inspect spacex-api

```text
NAME
  pm connectors inspect spacex-api - SpaceX API connector manual

SYNOPSIS
  pm connectors inspect spacex-api
  pm connectors inspect spacex-api --json
  pm credentials add <name> --connector spacex-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public SpaceX launch, rocket, core, capsule, crew, Dragon, history, payload, and Starlink data.

ICON
  id: spacex
  asset: icons/spacex.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://github.com/r-spacex/SpaceX-API/tree/master/docs

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  mode

ETL STREAMS
  launches:
    primary key: id
    fields: capsules(array), crew(array), date_local(string), date_precision(string), date_unix(integer), date_utc(string), details(string), flight_number(integer), id(string), launchpad(string), links(object), name(string), payloads(array), rocket(string), ships(array), success(boolean), upcoming(boolean)
  rockets:
    primary key: id
    fields: active(boolean), boosters(integer), company(string), cost_per_launch(integer), country(string), description(string), first_flight(string), id(string), name(string), stages(integer), success_rate_pct(integer), type(string)
  capsules:
    primary key: id
    fields: id(string), land_landings(integer), last_update(string), launches(array), reuse_count(integer), serial(string), status(string), type(string), water_landings(integer)
  cores:
    primary key: id
    fields: asds_attempts(integer), asds_landings(integer), block(integer), id(string), last_update(string), launches(array), reuse_count(integer), rtls_attempts(integer), rtls_landings(integer), serial(string), status(string)
  crew:
    primary key: id
    fields: agency(string), id(string), image(string), launches(array), name(string), status(string), wikipedia(string)
  dragons:
    primary key: id
    fields: active(boolean), crew_capacity(integer), description(string), dry_mass_kg(integer), first_flight(string), id(string), name(string), type(string)
  history:
    primary key: id
    fields: details(string), event_date_unix(integer), event_date_utc(string), id(string), links(object), title(string)
  payloads:
    primary key: id
    fields: customers(array), id(string), launch(string), manufacturers(array), mass_kg(number), name(string), nationalities(array), orbit(string), reused(boolean), type(string)
  starlink:
    primary key: id
    fields: height_km(number), id(string), latitude(number), launch(string), longitude(number), spaceTrack(object), velocity_kms(number), version(string)
  launchpads:
    primary key: id
    fields: full_name(string), id(string), latitude(number), launch_attempts(integer), launch_successes(integer), launches(array), locality(string), longitude(number), name(string), region(string), status(string)
  landpads:
    primary key: id
    fields: full_name(string), id(string), landing_attempts(integer), landing_successes(integer), latitude(number), launches(array), locality(string), longitude(number), name(string), region(string), status(string), type(string)
  ships:
    primary key: id
    fields: active(boolean), home_port(string), id(string), launches(array), name(string), roles(array), type(string), year_built(integer)
  roadster:
    primary key: id
    fields: earth_distance_km(number), id(string), launch_date_utc(string), launch_mass_kg(number), mars_distance_km(number), name(string), speed_kph(number), wikipedia(string)
  company:
    primary key: name
    fields: ceo(string), cto(string), employees(integer), founded(integer), founder(string), headquarters(object), name(string), summary(string), valuation(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external public SpaceX API read of launch and vehicle data
  approval: none; read-only public API, no credentials
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect spacex-api

  # Inspect as structured JSON
  pm connectors inspect spacex-api --json

AGENT WORKFLOW
  - Run pm connectors inspect spacex-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
