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
    fields: capsules(), crew(), date_local(), date_precision(), date_unix(), date_utc(), details(), flight_number(), id(), launchpad(), links(), name(), payloads(), rocket(), ships(), success(), upcoming()
  rockets:
    primary key: id
    fields: active(), boosters(), company(), cost_per_launch(), country(), description(), first_flight(), id(), name(), stages(), success_rate_pct(), type()
  capsules:
    primary key: id
    fields: id(), land_landings(), last_update(), launches(), reuse_count(), serial(), status(), type(), water_landings()
  cores:
    primary key: id
    fields: asds_attempts(), asds_landings(), block(), id(), last_update(), launches(), reuse_count(), rtls_attempts(), rtls_landings(), serial(), status()
  crew:
    primary key: id
    fields: agency(), id(), image(), launches(), name(), status(), wikipedia()
  dragons:
    primary key: id
    fields: active(), crew_capacity(), description(), dry_mass_kg(), first_flight(), id(), name(), type()
  history:
    primary key: id
    fields: details(), event_date_unix(), event_date_utc(), id(), links(), title()
  payloads:
    primary key: id
    fields: customers(), id(), launch(), manufacturers(), mass_kg(), name(), nationalities(), orbit(), reused(), type()
  starlink:
    primary key: id
    fields: height_km(), id(), latitude(), launch(), longitude(), spaceTrack(), velocity_kms(), version()
  launchpads:
    primary key: id
    fields: full_name(), id(), latitude(), launch_attempts(), launch_successes(), launches(), locality(), longitude(), name(), region(), status()
  landpads:
    primary key: id
    fields: full_name(), id(), landing_attempts(), landing_successes(), latitude(), launches(), locality(), longitude(), name(), region(), status(), type()
  ships:
    primary key: id
    fields: active(), home_port(), id(), launches(), name(), roles(), type(), year_built()
  roadster:
    primary key: id
    fields: earth_distance_km(), id(), launch_date_utc(), launch_mass_kg(), mars_distance_km(), name(), speed_kph(), wikipedia()
  company:
    primary key: name
    fields: ceo(), cto(), employees(), founded(), founder(), headquarters(), name(), summary(), valuation()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
