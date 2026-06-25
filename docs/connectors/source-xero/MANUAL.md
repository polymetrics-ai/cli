# pm connectors inspect source-xero

```text
NAME
  pm connectors inspect source-xero - Xero connector manual

SYNOPSIS
  pm connectors inspect source-xero
  pm connectors inspect source-xero --json
  pm credentials add <name> --connector source-xero [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Xero catalog connector for https://docs.airbyte.com/integrations/sources/xero. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-xero:2.1.5 (metadata only; not executed)

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
  Xero API reference: https://developer.xero.com/documentation/
  Xero authentication: https://developer.xero.com/documentation/guides/oauth2/overview/
  Xero rate limits: https://developer.xero.com/documentation/guides/oauth2/limits/
  Xero Status: https://status.xero.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/xero

CONFIGURATION
  credentials (object) required
  start_date (string) required: UTC date and time in the format YYYY-MM-DDTHH:mm:ssZ. Any data with created_at before this date will not be synced.
  tenant_id (string) required secret: Enter your Xero organization's Tenant ID
  secret fields: credentials.access_token, credentials.client_secret, tenant_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/xero

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-xero

  # Inspect as JSON
  pm connectors inspect source-xero --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Xero documentation: https://docs.airbyte.com/integrations/sources/xero

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
