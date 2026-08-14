# pm connectors inspect care-quality-commission

```text
NAME
  pm connectors inspect care-quality-commission - Care Quality Commission connector manual

SYNOPSIS
  pm connectors inspect care-quality-commission
  pm connectors inspect care-quality-commission --json
  pm credentials add <name> --connector care-quality-commission [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Care Quality Commission (CQC) registered locations, providers, and inspection areas from the public CQC Syndication API. Read-only (full-refresh).

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
  mode
  api_key (secret) (required)

ETL STREAMS
  locations:
    primary key: locationId
    fields: locationId(string), locationName(string), postalCode(string)
  providers:
    primary key: providerId
    fields: providerId(string), providerName(string)
  inspection_areas:
    primary key: inspectionAreaId
    fields: endDate(string), inspectionAreaId(string), inspectionAreaName(string), inspectionAreaType(string), inspectionCategories(array), orgInspectionAreaRetirementDate(string), status(string), supersededBy(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external CQC Syndication API read of publicly published care provider/location data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect care-quality-commission

  # Inspect as structured JSON
  pm connectors inspect care-quality-commission --json

AGENT WORKFLOW
  - Run pm connectors inspect care-quality-commission before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
