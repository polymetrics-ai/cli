# pm connectors inspect source-amplitude

```text
NAME
  pm connectors inspect source-amplitude - Amplitude connector manual

SYNOPSIS
  pm connectors inspect source-amplitude
  pm connectors inspect source-amplitude --json
  pm credentials add <name> --connector source-amplitude [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Amplitude catalog connector for https://docs.airbyte.com/integrations/sources/amplitude. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-amplitude:0.7.32 (metadata only; not executed)

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
  Analytics API: https://www.docs.developers.amplitude.com/analytics/apis/http-v2-api/
  Authentication: https://www.docs.developers.amplitude.com/analytics/apis/authentication/
  Rate limits: https://www.docs.developers.amplitude.com/analytics/apis/http-v2-api/#rate-limits
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/amplitude

CONFIGURATION
  active_users_group_by_country (boolean): According to <a href="https://amplitude.com/docs/apis/analytics/dashboard-rest#query-parameters">Amplitude documentation</a>, grouping by `Country` is optional. If you face issu...
  api_key (string) required secret: Amplitude API Key. See the <a href="https://docs.airbyte.com/integrations/sources/amplitude#setup-guide">setup guide</a> for more information on how to obtain this key.
  data_region (string): Amplitude data region server
  request_time_range (integer): According to <a href="https://www.docs.developers.amplitude.com/analytics/apis/export-api/#considerations">Considerations</a> too large of a time range in te request can cause a...
  secret_key (string) required secret: Amplitude Secret Key. See the <a href="https://docs.airbyte.com/integrations/sources/amplitude#setup-guide">setup guide</a> for more information on how to obtain this key.
  start_date (string) required: UTC date and time in the format 2021-01-25T00:00:00Z. Any data before this date will not be replicated.
  secret fields: api_key, secret_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/amplitude

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-amplitude

  # Inspect as JSON
  pm connectors inspect source-amplitude --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Amplitude documentation: https://docs.airbyte.com/integrations/sources/amplitude

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
