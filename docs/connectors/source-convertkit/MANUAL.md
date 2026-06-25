# pm connectors inspect source-convertkit

```text
NAME
  pm connectors inspect source-convertkit - ConvertKit connector manual

SYNOPSIS
  pm connectors inspect source-convertkit
  pm connectors inspect source-convertkit --json
  pm credentials add <name> --connector source-convertkit [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  ConvertKit catalog connector for https://docs.airbyte.com/integrations/sources/convertkit. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-convertkit:0.3.3 (metadata only; not executed)

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
  ConvertKit API reference: https://developers.convertkit.com/
  ConvertKit authentication: https://developers.convertkit.com/#authentication
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/convertkit

CONFIGURATION
  api_secret (string)
  credentials (object) required
  start_date (string) required
  secret fields: credentials.access_token, credentials.api_key, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/convertkit

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-convertkit

  # Inspect as JSON
  pm connectors inspect source-convertkit --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  ConvertKit documentation: https://docs.airbyte.com/integrations/sources/convertkit

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
