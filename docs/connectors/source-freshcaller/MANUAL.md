# pm connectors inspect source-freshcaller

```text
NAME
  pm connectors inspect source-freshcaller - Freshcaller connector manual

SYNOPSIS
  pm connectors inspect source-freshcaller
  pm connectors inspect source-freshcaller --json
  pm credentials add <name> --connector source-freshcaller [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Freshcaller catalog connector for https://docs.airbyte.com/integrations/sources/freshcaller. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-freshcaller:0.5.4 (metadata only; not executed)

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
  Freshcaller API reference: https://developers.freshcaller.com/api/
  Freshcaller authentication: https://developers.freshcaller.com/api/#authentication
  Freshworks Status: https://status.freshworks.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/freshcaller

CONFIGURATION
  api_key (string) required secret: Freshcaller API Key. See the docs for more information on how to obtain this key.
  domain (string) required: Used to construct Base URL for the Freshcaller APIs
  requests_per_minute (integer): The number of requests per minute that this source allowed to use. There is a rate limit of 50 requests per minute per app per account.
  start_date (string): UTC date and time. Any data created after this date will be replicated.
  sync_lag_minutes (integer): Lag in minutes for each sync, i.e., at time T, data for the time range [prev_sync_time, T-30] will be fetched
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/freshcaller

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-freshcaller

  # Inspect as JSON
  pm connectors inspect source-freshcaller --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Freshcaller documentation: https://docs.airbyte.com/integrations/sources/freshcaller

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
