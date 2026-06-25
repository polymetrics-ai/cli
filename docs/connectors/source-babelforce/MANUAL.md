# pm connectors inspect source-babelforce

```text
NAME
  pm connectors inspect source-babelforce - Babelforce connector manual

SYNOPSIS
  pm connectors inspect source-babelforce
  pm connectors inspect source-babelforce --json
  pm credentials add <name> --connector source-babelforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Babelforce catalog connector for https://docs.airbyte.com/integrations/sources/babelforce. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-babelforce:0.3.30 (metadata only; not executed)

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
  API documentation: https://api.babelforce.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/babelforce

CONFIGURATION
  access_key_id (string) required secret: The Babelforce access key ID
  access_token (string) required secret: The Babelforce access token
  date_created_from (integer): Timestamp in Unix the replication from Babelforce API will start from. For example 1651363200 which corresponds to 2022-05-01 00:00:00.
  date_created_to (integer): Timestamp in Unix the replication from Babelforce will be up to. For example 1651363200 which corresponds to 2022-05-01 00:00:00.
  region (string) required: Babelforce region
  secret fields: access_key_id, access_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/babelforce

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-babelforce

  # Inspect as JSON
  pm connectors inspect source-babelforce --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Babelforce documentation: https://docs.airbyte.com/integrations/sources/babelforce

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
