# pm connectors inspect source-tiktok-marketing

```text
NAME
  pm connectors inspect source-tiktok-marketing - TikTok Marketing connector manual

SYNOPSIS
  pm connectors inspect source-tiktok-marketing
  pm connectors inspect source-tiktok-marketing --json
  pm credentials add <name> --connector source-tiktok-marketing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  TikTok Marketing catalog connector for https://docs.airbyte.com/integrations/sources/tiktok-marketing. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-tiktok-marketing:5.1.1 (metadata only; not executed)

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
  Versioning docs: https://business-api.tiktok.com/portal/docs?id=1740029169927169
  Changelog: https://business-api.tiktok.com/portal/docs?id=1740029165513730
  TikTok Business API Documentation: https://business-api.tiktok.com/portal/docs?id=1740302848670722
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/tiktok-marketing

CONFIGURATION
  attribution_window (integer): The attribution window in days.
  credentials (object): Authentication method
  end_date (string): The date until which you'd like to replicate data for all incremental streams, in the format YYYY-MM-DD. All data generated between start_date and this date will be replicated. ...
  include_deleted (boolean): Set to active if you want to include deleted data in report based streams and Ads, Ad Groups and Campaign streams.
  report_granularity (integer): The number of days per API request for daily report streams. Use the default 30 for most accounts. If syncs fail with TikTok API error 40067 ("query too large"), reduce this val...
  start_date (string): The Start Date in format: YYYY-MM-DD. Any data before this date will not be replicated. If this parameter is not set, all data will be replicated.
  secret fields: credentials.access_token, credentials.app_id, credentials.secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/tiktok-marketing

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-tiktok-marketing

  # Inspect as JSON
  pm connectors inspect source-tiktok-marketing --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  TikTok Marketing documentation: https://docs.airbyte.com/integrations/sources/tiktok-marketing

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
