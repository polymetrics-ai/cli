# pm connectors inspect source-freshdesk

```text
NAME
  pm connectors inspect source-freshdesk - Freshdesk connector manual

SYNOPSIS
  pm connectors inspect source-freshdesk
  pm connectors inspect source-freshdesk --json
  pm credentials add <name> --connector source-freshdesk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Freshdesk catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/freshdesk.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.freshdesk.com/api/#change_log

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

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
  Changelog: https://developers.freshdesk.com/api/#change_log

CONFIGURATION
  api_key (string) required secret: manual intervention needed
  domain (string) required: Freshdesk domain
  lookback_window_in_days (integer): Number of days for lookback window for the stream Satisfaction Ratings
  num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may increase API rate limit usage. Adjust based on your Freshdesk API plan.
  rate_limit_plan (object): Rate Limit Plan for API Budget
  requests_per_minute (integer): The number of requests per minute that this source allowed to use. There is a rate limit of 50 requests per minute per app per account.
  start_date (string): UTC date and time. Any data created after this date will be replicated. If this parameter is not set, all data will be replicated.
  subscription_tier (string): Your API subscription tier (affects rate limits)
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-freshdesk

  # Inspect as JSON
  pm connectors inspect source-freshdesk --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Changelog: https://developers.freshdesk.com/api/#change_log

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
