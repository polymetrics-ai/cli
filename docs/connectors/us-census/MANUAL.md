# pm connectors inspect us-census

```text
NAME
  pm connectors inspect us-census - US Census connector manual

SYNOPSIS
  pm connectors inspect us-census
  pm connectors inspect us-census --json
  pm credentials add <name> --connector us-census [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads configured datasets from the US Census Bureau's API via a caller-supplied query path and query-string qualifier, and reads the Bureau's own published dataset catalog.

ICON
  asset: icons/uscensus.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.census.gov/data/developers/data-sets.html

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  query_params
  query_path
  api_key (secret)

ETL STREAMS
  query:
    primary key: name
    fields: estab(), name()
  datasets:
    primary key: identifier
    fields: accessLevel(), c_dataset(), c_geographyLink(), c_isAvailable(), c_variablesLink(), c_vintage(), dataset_path(), description(), identifier(), modified(), title()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external US Census Bureau API read of a caller-configured dataset endpoint, plus the Bureau's own public dataset catalog (no auth required for the catalog)
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect us-census

  # Inspect as structured JSON
  pm connectors inspect us-census --json

AGENT WORKFLOW
  - Run pm connectors inspect us-census before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
