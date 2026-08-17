# pm connectors inspect jamf-pro

```text
NAME
  pm connectors inspect jamf-pro - Jamf Pro connector manual

SYNOPSIS
  pm connectors inspect jamf-pro
  pm connectors inspect jamf-pro --json
  pm credentials add <name> --connector jamf-pro [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Jamf Pro buildings, departments, categories, and scripts through the Jamf Pro REST API using Basic-credential token-exchange authentication.

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
  max_pages
  mode
  page_size
  username
  password (secret)

ETL STREAMS
  buildings:
    primary key: id
    fields: city(), country(), id(), name(), stateProvince(), streetAddress1(), streetAddress2(), zipPostalCode()
  departments:
    primary key: id
    fields: id(), name()
  categories:
    primary key: id
    fields: id(), name(), priority()
  scripts:
    primary key: id
    fields: categoryId(), categoryName(), id(), info(), name(), notes(), osRequirements(), priority()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Jamf Pro API read of MDM configuration data (buildings, departments, categories, scripts)
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect jamf-pro

  # Inspect as structured JSON
  pm connectors inspect jamf-pro --json

AGENT WORKFLOW
  - Run pm connectors inspect jamf-pro before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
