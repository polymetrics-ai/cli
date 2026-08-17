# pm connectors inspect open-data-dc

```text
NAME
  pm connectors inspect open-data-dc - Open Data DC connector manual

SYNOPSIS
  pm connectors inspect open-data-dc
  pm connectors inspect open-data-dc --json
  pm credentials add <name> --connector open-data-dc [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads District of Columbia Master Address Repository (MAR 2) locations, units, and SSL parcel records via the Open Data DC API. Read-only.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  location
  marid
  mode
  api_key (secret)

ETL STREAMS
  locations:
    primary key: MarId
    fields: AddrNum(), Anc(), CensusTract(), FullAddress(), Latitude(), Longitude(), MarId(), Quadrant(), ResidenceType(), SSL(), StName(), Status(), Ward(), Xcoord(), Ycoord(), Zipcode(), distance()
  units:
    primary key: UnitNum
    fields: FullAddress(), MarId(), Status(), UnitNum(), UnitSSL(), UnitType()
  ssls:
    primary key: SSL
    fields: Col(), FullAddress(), Lot(), Lot_type(), MarId(), SSL(), Square()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Open Data DC (MAR 2) API read of public address/parcel data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect open-data-dc

  # Inspect as structured JSON
  pm connectors inspect open-data-dc --json

AGENT WORKFLOW
  - Run pm connectors inspect open-data-dc before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
