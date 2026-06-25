# pm connectors inspect source-luma

```text
NAME
  pm connectors inspect source-luma - lu.ma connector manual

SYNOPSIS
  pm connectors inspect source-luma
  pm connectors inspect source-luma --json
  pm credentials add <name> --connector source-luma [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  lu.ma catalog connector for https://docs.airbyte.com/integrations/sources/luma. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-luma:0.0.61 (metadata only; not executed)

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
  Luma API documentation: https://docs.lu.ma/reference/getting-started
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/luma

CONFIGURATION
  api_key (string) required secret: Get your API key on lu.ma Calendars dashboard → Settings.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/luma

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-luma

  # Inspect as JSON
  pm connectors inspect source-luma --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  lu.ma documentation: https://docs.airbyte.com/integrations/sources/luma

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
