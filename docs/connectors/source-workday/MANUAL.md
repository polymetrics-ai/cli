# pm connectors inspect source-workday

```text
NAME
  pm connectors inspect source-workday - Workday connector manual

SYNOPSIS
  pm connectors inspect source-workday
  pm connectors inspect source-workday --json
  pm credentials add <name> --connector source-workday [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Workday catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/workday.svg
  source: official
  review_status: official_verified
  review_url: https://community.workday.com/sites/default/files/file-hosting/productionapi/index.html

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: certified

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
  Workday documentation: https://community.workday.com/sites/default/files/file-hosting/productionapi/index.html

CONFIGURATION
  credentials (object) required: Credentials for connecting to the Workday (RAAS) API.
  host (string) required
  num_workers (integer): The number of worker threads to use for the sync.
  report_ids (array) required: Report IDs can be found by clicking the three dots on the right side of the report > Web Service > View URLs > in JSON url copy everything between Workday tenant/ and ?format=json.
  tenant_id (string) required secret
  secret fields: credentials.password, tenant_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-workday

  # Inspect as JSON
  pm connectors inspect source-workday --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Workday documentation: https://community.workday.com/sites/default/files/file-hosting/productionapi/index.html

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
