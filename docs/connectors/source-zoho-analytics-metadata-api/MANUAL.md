# pm connectors inspect source-zoho-analytics-metadata-api

```text
NAME
  pm connectors inspect source-zoho-analytics-metadata-api - Zoho Analytics metadata api connector manual

SYNOPSIS
  pm connectors inspect source-zoho-analytics-metadata-api
  pm connectors inspect source-zoho-analytics-metadata-api --json
  pm credentials add <name> --connector source-zoho-analytics-metadata-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Zoho Analytics metadata api catalog connector for https://docs.airbyte.com/integrations/sources/zoho-analytics-metadata-api. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-zoho-analytics-metadata-api:0.0.56 (metadata only; not executed)

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
  Zoho Analytics Metadata API: https://www.zoho.com/analytics/api/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/zoho-analytics-metadata-api

CONFIGURATION
  client_id (string) required secret
  client_secret (string) required secret
  data_center (string) required
  org_id (number) required
  refresh_token (string) required secret
  secret fields: client_id, client_secret, refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/zoho-analytics-metadata-api

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-zoho-analytics-metadata-api

  # Inspect as JSON
  pm connectors inspect source-zoho-analytics-metadata-api --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Zoho Analytics metadata api documentation: https://docs.airbyte.com/integrations/sources/zoho-analytics-metadata-api

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
