# pm connectors inspect source-my-hours

```text
NAME
  pm connectors inspect source-my-hours - My Hours connector manual

SYNOPSIS
  pm connectors inspect source-my-hours
  pm connectors inspect source-my-hours --json
  pm credentials add <name> --connector source-my-hours [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  My Hours catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/my-hours.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://myhours.com/api

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
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
  family: declarative_http_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  My Hours API documentation: https://myhours.com/api

CONFIGURATION
  email (string) required: Your My Hours username
  logs_batch_size (integer): Pagination size used for retrieving logs in days
  password (string) required secret: The password associated to the username
  start_date (string) required: Start date for collecting time logs
  secret fields: password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-my-hours

  # Inspect as JSON
  pm connectors inspect source-my-hours --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  My Hours API documentation: https://myhours.com/api

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
