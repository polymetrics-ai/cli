# pm connectors inspect source-spacex-api

```text
NAME
  pm connectors inspect source-spacex-api - SpaceX API connector manual

SYNOPSIS
  pm connectors inspect source-spacex-api
  pm connectors inspect source-spacex-api --json
  pm credentials add <name> --connector source-spacex-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SpaceX API catalog connector for https://docs.airbyte.com/integrations/sources/spacex-api. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-spacex-api:0.2.22 (metadata only; not executed)

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
  SpaceX API documentation: https://github.com/r-spacex/SpaceX-API/tree/master/docs
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/spacex-api

CONFIGURATION
  id (string)
  options (string)

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/spacex-api

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-spacex-api

  # Inspect as JSON
  pm connectors inspect source-spacex-api --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SpaceX API documentation: https://docs.airbyte.com/integrations/sources/spacex-api

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
