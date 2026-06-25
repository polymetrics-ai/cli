# pm connectors inspect source-linear

```text
NAME
  pm connectors inspect source-linear - Linear connector manual

SYNOPSIS
  pm connectors inspect source-linear
  pm connectors inspect source-linear --json
  pm credentials add <name> --connector source-linear [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Linear catalog connector for https://docs.airbyte.com/integrations/sources/linear. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-linear:0.2.7 (metadata only; not executed)

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
  Linear API reference: https://developers.linear.app/docs/graphql/working-with-the-graphql-api
  Linear authentication: https://developers.linear.app/docs/oauth/authentication
  Linear rate limits: https://developers.linear.app/docs/graphql/working-with-the-graphql-api#rate-limiting
  Linear Status: https://status.linear.app/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/linear

CONFIGURATION
  credentials (object) required: Choose how to authenticate to Linear.
  num_workers (integer): Number of worker threads used to read streams in parallel. Higher values can speed up syncs but increase the risk of hitting Linear's rate limits. The Linear API limits per-user...
  start_date (string): UTC date and time in the ISO 8601 format (e.g. 2020-01-01T00:00:00.000Z). Any records updated before this date will not be replicated. Only applies to streams that support incre...
  secret fields: credentials.api_key, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/linear

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-linear

  # Inspect as JSON
  pm connectors inspect source-linear --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Linear documentation: https://docs.airbyte.com/integrations/sources/linear

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
