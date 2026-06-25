# pm connectors inspect source-appsflyer

```text
NAME
  pm connectors inspect source-appsflyer - AppsFlyer connector manual

SYNOPSIS
  pm connectors inspect source-appsflyer
  pm connectors inspect source-appsflyer --json
  pm credentials add <name> --connector source-appsflyer [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  AppsFlyer catalog connector for https://docs.airbyte.com/integrations/sources/appsflyer. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-appsflyer:0.3.0 (metadata only; not executed)

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
  API documentation: https://dev.appsflyer.com/hc/reference
  Authentication: https://dev.appsflyer.com/hc/docs/authentication
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/appsflyer

CONFIGURATION
  api_token (string) required secret: Pull API token for authentication. If you change the account admin, the token changes, and you must update scripts with the new token. <a href="https://support.appsflyer.com/hc/...
  app_id (string) required: App identifier as found in AppsFlyer.
  start_date (string) required: The default value to use if no bookmark exists for an endpoint. Raw Reports historical lookback is limited to 90 days.
  timezone (string): Time zone in which date times are stored. The project timezone may be found in the App settings in the AppsFlyer console.
  secret fields: api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/appsflyer

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-appsflyer

  # Inspect as JSON
  pm connectors inspect source-appsflyer --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  AppsFlyer documentation: https://docs.airbyte.com/integrations/sources/appsflyer

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
