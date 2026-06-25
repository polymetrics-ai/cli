# pm connectors inspect source-tempo

```text
NAME
  pm connectors inspect source-tempo - Tempo connector manual

SYNOPSIS
  pm connectors inspect source-tempo
  pm connectors inspect source-tempo --json
  pm credentials add <name> --connector source-tempo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Tempo catalog connector for https://docs.airbyte.com/integrations/sources/tempo. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-tempo:0.4.52 (metadata only; not executed)

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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Tempo Timesheets API: https://apidocs.tempo.io/
  Tempo authentication: https://apidocs.tempo.io/#section/Authentication
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/tempo

CONFIGURATION
  api_token (string) required secret: Tempo API Token. Go to Tempo>Settings, scroll down to Data Access and select API integration.
  secret fields: api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/tempo

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-tempo

  # Inspect as JSON
  pm connectors inspect source-tempo --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Tempo documentation: https://docs.airbyte.com/integrations/sources/tempo

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
