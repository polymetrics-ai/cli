# pm connectors inspect source-fullstory

```text
NAME
  pm connectors inspect source-fullstory - Fullstory connector manual

SYNOPSIS
  pm connectors inspect source-fullstory
  pm connectors inspect source-fullstory --json
  pm credentials add <name> --connector source-fullstory [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Fullstory catalog connector for https://docs.airbyte.com/integrations/sources/fullstory. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-fullstory:0.2.23 (metadata only; not executed)

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
  FullStory API reference: https://developer.fullstory.com/reference
  FullStory authentication: https://developer.fullstory.com/introduction
  FullStory Status: https://status.fullstory.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/fullstory

CONFIGURATION
  api_key (string) required secret: API Key for the fullstory.com API.
  uid (string) required secret: User ID for the fullstory.com API.
  secret fields: api_key, uid

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/fullstory

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-fullstory

  # Inspect as JSON
  pm connectors inspect source-fullstory --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Fullstory documentation: https://docs.airbyte.com/integrations/sources/fullstory

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
