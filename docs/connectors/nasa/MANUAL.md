# pm connectors inspect nasa

```text
NAME
  pm connectors inspect nasa - NASA connector manual

SYNOPSIS
  pm connectors inspect nasa
  pm connectors inspect nasa --json
  pm credentials add <name> --connector nasa [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads NASA Open API data: Astronomy Picture of the Day, Near-Earth Objects (NeoWs feed and browse), EPIC Earth imagery, and Mars rover photos. Read-only.

ICON
  asset: icons/nasa.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.nasa.gov/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  count
  end_date
  mode
  sol
  start_date
  thumbs
  api_key (secret)

ETL STREAMS
  apod:
    primary key: date
    cursor: date
    fields: copyright(), date(), explanation(), hdurl(), media_type(), service_version(), thumbnail_url(), title(), url()
  neo_feed:
    primary key: id
    fields: absolute_magnitude_h(), id(), is_potentially_hazardous_asteroid(), is_sentry_object(), name(), nasa_jpl_url(), neo_reference_id()
  neo_browse:
    primary key: id
    fields: absolute_magnitude_h(), id(), is_potentially_hazardous_asteroid(), is_sentry_object(), name(), nasa_jpl_url(), neo_reference_id()
  epic:
    primary key: identifier
    fields: caption(), date(), identifier(), image(), version()
  mars_photos:
    primary key: id
    fields: camera(), earth_date(), id(), img_src(), rover(), sol()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external NASA Open API read of public astronomy and space data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nasa

  # Inspect as structured JSON
  pm connectors inspect nasa --json

AGENT WORKFLOW
  - Run pm connectors inspect nasa before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
