# pm connectors inspect source-google-ads

```text
NAME
  pm connectors inspect source-google-ads - Google Ads connector manual

SYNOPSIS
  pm connectors inspect source-google-ads
  pm connectors inspect source-google-ads --json
  pm credentials add <name> --connector source-google-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google Ads catalog connector for https://docs.airbyte.com/integrations/sources/google-ads. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-google-ads:6.0.0 (metadata only; not executed)

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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Release notes: https://developers.google.com/google-ads/api/docs/release-notes
  Developer blog: https://ads-developers.googleblog.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/google-ads

CONFIGURATION
  conversion_window_days (integer): A conversion window is the number of days after an ad interaction (such as an ad click or video view) during which a conversion, such as a purchase, is recorded in Google Ads. F...
  credentials (object) required
  custom_queries_array (array)
  customer_id (string): Comma-separated list of (client) customer IDs. Each customer ID must be specified as a 10-digit number without dashes. For detailed instructions on finding this value, refer to ...
  customer_status_filter (array): A list of customer statuses to filter on. For detailed info about what each status mean refer to Google Ads <a href="https://developers.google.com/google-ads/api/reference/rpc/v...
  end_date (string): UTC date in the format YYYY-MM-DD. Any data after this date will not be replicated. (Default value of today is used if not set)
  num_workers (integer): The number of concurrent threads to use for syncing. Increasing this value may speed up syncs for accounts with many customers or streams. Adjust based on your API usage and rat...
  start_date (string): UTC date in the format YYYY-MM-DD. Any data before this date will not be replicated. (Default value of two years ago is used if not set)
  secret fields: credentials.access_token, credentials.client_secret, credentials.developer_token, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/google-ads

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-google-ads

  # Inspect as JSON
  pm connectors inspect source-google-ads --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google Ads documentation: https://docs.airbyte.com/integrations/sources/google-ads

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
