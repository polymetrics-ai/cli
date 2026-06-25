# pm connectors inspect source-bing-ads

```text
NAME
  pm connectors inspect source-bing-ads - Bing Ads connector manual

SYNOPSIS
  pm connectors inspect source-bing-ads
  pm connectors inspect source-bing-ads --json
  pm credentials add <name> --connector source-bing-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Bing Ads catalog connector for https://docs.airbyte.com/integrations/sources/bing-ads. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-bing-ads:3.0.0 (metadata only; not executed)

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
  Bing Ads API Release Notes: https://learn.microsoft.com/en-us/advertising/guides/release-notes
  Release notes: https://learn.microsoft.com/en-us/advertising/guides/release-notes?view=bingads-13
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/bing-ads

CONFIGURATION
  account_names (array): Predicates that will be used to sync data by specific accounts.
  auth_method (string)
  client_id (string) required secret: The Client ID of your Microsoft Advertising developer application.
  client_secret (string) secret: The Client Secret of your Microsoft Advertising developer application.
  custom_reports (array): You can add your Custom Bing Ads report by creating one.
  developer_token (string) required secret: Developer token associated with user. See more info <a href="https://docs.microsoft.com/en-us/advertising/guides/get-started?view=bingads-13#get-developer-token"> in the docs</a>.
  lookback_window (integer): Also known as attribution or conversion window. How far into the past to look for records (in days). If your conversion window has an hours/minutes granularity, round it up to t...
  num_workers (integer): The number of worker threads to use for the sync. Increase this to speed up syncs for accounts with many reports. The default should work for most use cases.
  refresh_token (string) required secret: Refresh Token to renew the expired Access Token.
  reports_start_date (string): The start date from which to begin replicating report data. Any data generated before this date will not be replicated in reports. This is a UTC date in YYYY-MM-DD format. If no...
  tenant_id (string) secret: The Tenant ID of your Microsoft Advertising developer application. Set this to "common" unless you know you need a different value.
  secret fields: client_id, client_secret, developer_token, refresh_token, tenant_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/bing-ads

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-bing-ads

  # Inspect as JSON
  pm connectors inspect source-bing-ads --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Bing Ads documentation: https://docs.airbyte.com/integrations/sources/bing-ads

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
