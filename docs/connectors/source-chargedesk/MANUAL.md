# pm connectors inspect source-chargedesk

```text
NAME
  pm connectors inspect source-chargedesk - Chargedesk connector manual

SYNOPSIS
  pm connectors inspect source-chargedesk
  pm connectors inspect source-chargedesk --json
  pm credentials add <name> --connector source-chargedesk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Chargedesk catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  Chargedesk API documentation: https://chargedesk.com/api-docs
  Chargedesk authentication: https://chargedesk.com/api-docs/authentication

CONFIGURATION
  password (string) secret
  start_date (integer): Date from when the sync should start in epoch Unix timestamp
  username (string) required
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
  pm connectors inspect source-chargedesk

  # Inspect as JSON
  pm connectors inspect source-chargedesk --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Chargedesk API documentation: https://chargedesk.com/api-docs
  Chargedesk authentication: https://chargedesk.com/api-docs/authentication

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
