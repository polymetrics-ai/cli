# pm connectors inspect source-genesys

```text
NAME
  pm connectors inspect source-genesys - Genesys connector manual

SYNOPSIS
  pm connectors inspect source-genesys
  pm connectors inspect source-genesys --json
  pm credentials add <name> --connector source-genesys [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Genesys catalog connector for https://docs.airbyte.com/integrations/sources/genesys. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-genesys:0.1.41 (metadata only; not executed)

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
  family: custom_go_port
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Genesys Cloud API reference: https://developer.genesys.cloud/api/
  Genesys authentication: https://developer.genesys.cloud/authorization/
  Genesys rate limits: https://developer.genesys.cloud/api/rest/rate_limits
  Genesys Cloud Status: https://status.mypurecloud.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/genesys

CONFIGURATION
  client_id (string) required secret: Your OAuth user Client ID
  client_secret (string) required secret: Your OAuth user Client Secret
  start_date (string) required: Start Date in format: YYYY-MM-DD
  tenant_endpoint (string) required: Please choose the right endpoint where your Tenant is located. More info by this <a href="https://help.mypurecloud.com/articles/aws-regions-for-genesys-cloud-deployment/">Link</a>
  secret fields: client_id, client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/genesys

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-genesys

  # Inspect as JSON
  pm connectors inspect source-genesys --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Genesys documentation: https://docs.airbyte.com/integrations/sources/genesys

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
