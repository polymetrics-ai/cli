# pm connectors inspect source-monday

```text
NAME
  pm connectors inspect source-monday - Monday connector manual

SYNOPSIS
  pm connectors inspect source-monday
  pm connectors inspect source-monday --json
  pm credentials add <name> --connector source-monday [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Monday catalog connector for https://docs.airbyte.com/integrations/sources/monday. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-monday:2.5.11 (metadata only; not executed)

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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  monday.com API reference: https://developer.monday.com/api-reference/docs
  monday.com authentication: https://developer.monday.com/api-reference/docs/authentication
  monday.com rate limits: https://developer.monday.com/api-reference/docs/rate-limits
  monday.com Status: https://status.monday.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/monday

CONFIGURATION
  board_ids (array): The IDs of the boards that the Items and Boards streams will extract records from. When left empty, streams will extract records from all boards that exist within the account.
  credentials (object)
  num_workers (integer): The number of worker threads to use for the sync.
  secret fields: credentials.access_token, credentials.api_token, credentials.client_id, credentials.client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/monday

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-monday

  # Inspect as JSON
  pm connectors inspect source-monday --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Monday documentation: https://docs.airbyte.com/integrations/sources/monday

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
