# pm connectors inspect ninjaone-rmm

```text
NAME
  pm connectors inspect ninjaone-rmm - NinjaOne RMM connector manual

SYNOPSIS
  pm connectors inspect ninjaone-rmm
  pm connectors inspect ninjaone-rmm --json
  pm credentials add <name> --connector ninjaone-rmm [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads NinjaOne RMM organizations, devices, locations, activities, and policies through the NinjaOne v2 REST API.

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
  page_size
  api_key (secret) (required)

ETL STREAMS
  organizations:
    primary key: id
    fields: description(string), id(integer), name(string), node_approval_mode(string)
  devices:
    primary key: id
    fields: approval_status(string), dns_name(string), id(integer), location_id(integer), node_class(string), offline(boolean), organization_id(integer), system_name(string)
  locations:
    primary key: id
    fields: address(string), description(string), id(integer), name(string), organization_id(integer)
  activities:
    primary key: id
    cursor: activityTime
    fields: activityTime(number), activity_type(string), device_id(integer), id(integer), message(string), status(string)
  policies:
    primary key: id
    fields: description(string), id(integer), name(string), node_class(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external NinjaOne RMM API read of managed device and organization data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ninjaone-rmm

  # Inspect as structured JSON
  pm connectors inspect ninjaone-rmm --json

AGENT WORKFLOW
  - Run pm connectors inspect ninjaone-rmm before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
