# pm connectors inspect source-dixa

```text
NAME
  pm connectors inspect source-dixa - Dixa connector manual

SYNOPSIS
  pm connectors inspect source-dixa
  pm connectors inspect source-dixa --json
  pm credentials add <name> --connector source-dixa [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Dixa catalog connector for https://docs.airbyte.com/integrations/sources/dixa. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-dixa:0.4.21 (metadata only; not executed)

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
  Dixa API reference: https://docs.dixa.io/openapi/
  Dixa authentication: https://docs.dixa.io/openapi/dixa-api/#section/Authentication
  Dixa rate limits: https://docs.dixa.io/openapi/dixa-api/#section/Rate-limiting
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/dixa

CONFIGURATION
  api_token (string) required secret: Dixa API token
  batch_size (integer): Number of days to batch into one request. Max 31.
  start_date (string) required: The connector pulls records updated from this date onwards.
  secret fields: api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/dixa

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-dixa

  # Inspect as JSON
  pm connectors inspect source-dixa --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Dixa documentation: https://docs.airbyte.com/integrations/sources/dixa

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
