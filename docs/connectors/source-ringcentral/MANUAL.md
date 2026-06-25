# pm connectors inspect source-ringcentral

```text
NAME
  pm connectors inspect source-ringcentral - Ringcentral connector manual

SYNOPSIS
  pm connectors inspect source-ringcentral
  pm connectors inspect source-ringcentral --json
  pm credentials add <name> --connector source-ringcentral [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Ringcentral catalog connector for https://docs.airbyte.com/integrations/sources/ringcentral. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-ringcentral:0.2.22 (metadata only; not executed)

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
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/ringcentral

CONFIGURATION
  account_id (string) required secret: Could be seen at response to basic api call to an endpoint with ~ operator. Example- (https://platform.devtest.ringcentral.com/restapi/v1.0/account/~/extension/~/business-hours)
  auth_token (string) required secret: Token could be recieved by following instructions at https://developers.ringcentral.com/api-reference/authentication
  extension_id (string) required secret: Could be seen at response to basic api call to an endpoint with ~ operator. Example- (https://platform.devtest.ringcentral.com/restapi/v1.0/account/~/extension/~/business-hours)
  secret fields: account_id, auth_token, extension_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/ringcentral

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-ringcentral

  # Inspect as JSON
  pm connectors inspect source-ringcentral --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Ringcentral documentation: https://docs.airbyte.com/integrations/sources/ringcentral

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
