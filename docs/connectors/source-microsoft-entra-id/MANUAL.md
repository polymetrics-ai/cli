# pm connectors inspect source-microsoft-entra-id

```text
NAME
  pm connectors inspect source-microsoft-entra-id - Microsoft Entra Id connector manual

SYNOPSIS
  pm connectors inspect source-microsoft-entra-id
  pm connectors inspect source-microsoft-entra-id --json
  pm credentials add <name> --connector source-microsoft-entra-id [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Microsoft Entra Id catalog connector for https://docs.airbyte.com/integrations/sources/microsoft-entra-id. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-microsoft-entra-id:0.0.51 (metadata only; not executed)

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
  Microsoft Entra ID API: https://learn.microsoft.com/en-us/graph/api/resources/azure-ad-overview
  Microsoft Graph authentication: https://learn.microsoft.com/en-us/graph/auth/
  Microsoft Graph throttling: https://learn.microsoft.com/en-us/graph/throttling
  Microsoft 365 Status: https://status.office365.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/microsoft-entra-id

CONFIGURATION
  client_id (string) required secret
  client_secret (string) required secret
  tenant_id (string) required secret
  user_id (string) required secret
  secret fields: client_id, client_secret, tenant_id, user_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/microsoft-entra-id

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-microsoft-entra-id

  # Inspect as JSON
  pm connectors inspect source-microsoft-entra-id --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Microsoft Entra Id documentation: https://docs.airbyte.com/integrations/sources/microsoft-entra-id

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
