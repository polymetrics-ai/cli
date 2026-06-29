# pm connectors inspect source-datagen

```text
NAME
  pm connectors inspect source-datagen - End-to-End Testing (datagen) connector manual

SYNOPSIS
  pm connectors inspect source-datagen
  pm connectors inspect source-datagen --json
  pm credentials add <name> --connector source-datagen [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  End-to-End Testing (datagen) catalog connector. Native implementation status: planned_native_port.

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
  runtime_kind: database_go
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
  family: database_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

OFFICIAL APPLICATION DOCUMENTATION
  manual intervention needed

CONFIGURATION
  concurrency (integer): manual intervention needed
  flavor (object) required: Different patterns for generating data
  max_records (integer) required: The number of record messages to emit from this connector. Min 1. Max 100 billion.

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-datagen

  # Inspect as JSON
  pm connectors inspect source-datagen --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
