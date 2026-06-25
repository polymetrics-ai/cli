# pm connectors inspect source-chartmogul

```text
NAME
  pm connectors inspect source-chartmogul - Chartmogul connector manual

SYNOPSIS
  pm connectors inspect source-chartmogul
  pm connectors inspect source-chartmogul --json
  pm credentials add <name> --connector source-chartmogul [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Chartmogul catalog connector for https://docs.airbyte.com/integrations/sources/chartmogul. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-chartmogul:1.1.49 (metadata only; not executed)

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
  ChartMogul API reference: https://dev.chartmogul.com/reference
  ChartMogul authentication: https://dev.chartmogul.com/docs/authentication
  ChartMogul rate limits: https://dev.chartmogul.com/docs/rate-limits
  ChartMogul Status: https://status.chartmogul.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/chartmogul

CONFIGURATION
  api_key (string) required secret: Your Chartmogul API key. See <a href="https://help.chartmogul.com/hc/en-us/articles/4407796325906-Creating-and-Managing-API-keys#creating-an-api-key"> the docs </a> for info on ...
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. When feasible, any data before this date will not be replicated.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/chartmogul

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-chartmogul

  # Inspect as JSON
  pm connectors inspect source-chartmogul --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Chartmogul documentation: https://docs.airbyte.com/integrations/sources/chartmogul

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
