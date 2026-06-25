# pm connectors inspect source-sonar-cloud

```text
NAME
  pm connectors inspect source-sonar-cloud - Sonar Cloud connector manual

SYNOPSIS
  pm connectors inspect source-sonar-cloud
  pm connectors inspect source-sonar-cloud --json
  pm credentials add <name> --connector source-sonar-cloud [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Sonar Cloud catalog connector for https://docs.airbyte.com/integrations/sources/sonar-cloud. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-sonar-cloud:0.2.50 (metadata only; not executed)

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
  SonarCloud Web API: https://sonarcloud.io/web_api
  SonarCloud authentication: https://docs.sonarsource.com/sonarcloud/advanced-setup/user-accounts/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/sonar-cloud

CONFIGURATION
  component_keys (array) required: Comma-separated list of component keys.
  end_date (string): To retrieve issues created before the given date (inclusive).
  organization (string) required: Organization key. See <a href="https://docs.sonarcloud.io/appendices/project-information/#project-and-organization-keys">here</a>.
  start_date (string): To retrieve issues created after the given date (inclusive).
  user_token (string) required secret: Your User Token. See <a href="https://docs.sonarcloud.io/advanced-setup/user-accounts/">here</a>. The token is case sensitive.
  secret fields: user_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/sonar-cloud

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-sonar-cloud

  # Inspect as JSON
  pm connectors inspect source-sonar-cloud --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Sonar Cloud documentation: https://docs.airbyte.com/integrations/sources/sonar-cloud

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
