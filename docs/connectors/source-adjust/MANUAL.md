# pm connectors inspect source-adjust

```text
NAME
  pm connectors inspect source-adjust - Adjust connector manual

SYNOPSIS
  pm connectors inspect source-adjust
  pm connectors inspect source-adjust --json
  pm credentials add <name> --connector source-adjust [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Adjust catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/adjust.svg
  source: official
  review_status: official_verified
  review_url: https://dev.adjust.com/en/api/rs-api/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

RUNTIME CAPABILITIES
  metadata=true
  check=false
  catalog=false
  read=false
  write=false
  query=false
  etl=false
  reverse_etl=false
  unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

NATIVE PORT PLAN
  family: custom_go_port
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Adjust documentation: https://dev.adjust.com/en/api/rs-api/

CONFIGURATION
  additional_metrics (array): Metrics names that are not pre-defined, such as cohort metrics or app specific metrics.
  api_token (string) required secret: Adjust API key, see https://help.adjust.com/en/article/report-service-api-authentication
  dimensions (array) required: Dimensions allow a user to break down metrics into groups using one or several parameters. For example, the number of installs by date, country and network. See https://help.adj...
  ingest_start (string) required: Data ingest start date.
  metrics (array) required: Select at least one metric to query.
  until_today (boolean): Syncs data up until today. Useful when running daily incremental syncs, and duplicates are not desired.
  secret fields: api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-adjust

  # Inspect as JSON
  pm connectors inspect source-adjust --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Adjust documentation: https://dev.adjust.com/en/api/rs-api/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
