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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

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
  api_key (secret) (required)

ETL STREAMS
  locations:
    primary key: MarId
    fields: AddrNum(string), Anc(string), CensusTract(string), FullAddress(string), Latitude(number), Longitude(number), MarId(string), Quadrant(string), ResidenceType(string), SSL(string), StName(string), Status(string), Ward(string), Xcoord(number), Ycoord(number), Zipcode(string), distance(number)
  units:
    primary key: UnitNum
    fields: FullAddress(string), MarId(string), Status(string), UnitNum(string), UnitSSL(string), UnitType(string)
  ssls:
    primary key: SSL
    fields: Col(string), FullAddress(string), Lot(string), Lot_type(string), MarId(string), SSL(string), Square(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
