# pm connectors inspect source-recharge

```text
NAME
  pm connectors inspect source-recharge - Recharge connector manual

SYNOPSIS
  pm connectors inspect source-recharge
  pm connectors inspect source-recharge --json
  pm credentials add <name> --connector source-recharge [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Recharge catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/recharge.svg
  source: official
  review_status: official_verified
  review_url: https://docs.getrecharge.com/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Recharge documentation: https://docs.getrecharge.com/

CONFIGURATION
  access_token (string) required secret: manual intervention needed
  lookback_window_days (integer): Specifies how many days of historical data should be reloaded each time the recharge connector runs.
  start_date (string) required: The date from which you'd like to replicate data for Recharge API, in the format YYYY-MM-DDT00:00:00Z. Any data before this date will not be replicated.
  use_orders_deprecated_api (boolean): Define whether or not the `Orders` stream should use the deprecated `2021-01` API version, or use `2021-11`, otherwise.
  secret fields: access_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-recharge

  # Inspect as JSON
  pm connectors inspect source-recharge --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Recharge documentation: https://docs.getrecharge.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
