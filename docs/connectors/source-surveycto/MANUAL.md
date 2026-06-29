# pm connectors inspect source-surveycto

```text
NAME
  pm connectors inspect source-surveycto - SurveyCTO connector manual

SYNOPSIS
  pm connectors inspect source-surveycto
  pm connectors inspect source-surveycto --json
  pm credentials add <name> --connector source-surveycto [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SurveyCTO catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/surveycto.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html

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
  SurveyCTO API documentation: https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html

CONFIGURATION
  form_id (array) required: Unique identifier for one of your forms
  password (string) required secret: Password to authenticate into the SurveyCTO server
  server_name (string) required: The name of the SurveryCTO server
  start_date (string): initial date for survey cto
  username (string) required: Username to authenticate into the SurveyCTO server
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
  pm connectors inspect source-surveycto

  # Inspect as JSON
  pm connectors inspect source-surveycto --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SurveyCTO API documentation: https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
