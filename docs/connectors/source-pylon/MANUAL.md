# pm connectors inspect source-pylon

```text
NAME
  pm connectors inspect source-pylon - Pylon connector manual

SYNOPSIS
  pm connectors inspect source-pylon
  pm connectors inspect source-pylon --json
  pm credentials add <name> --connector source-pylon [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Pylon catalog connector for https://docs.airbyte.com/integrations/sources/pylon. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-pylon:0.0.12 (metadata only; not executed)

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
  Pylon API reference: https://docs.usepylon.com/pylon-docs/developer/api/api-reference
  Pylon authentication: https://docs.usepylon.com/pylon-docs/developer/api/authentication
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/pylon

CONFIGURATION
  api_token (string) required secret: Your Pylon API token. Generate one in the Pylon dashboard under Settings > API.
  start_date (string): The date from which you want to start syncing data, in UTC. Format: YYYY-MM-DDTHH:MM:SSZ. If not provided, defaults to 1 month ago.
  secret fields: api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/pylon

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-pylon

  # Inspect as JSON
  pm connectors inspect source-pylon --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Pylon documentation: https://docs.airbyte.com/integrations/sources/pylon

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
