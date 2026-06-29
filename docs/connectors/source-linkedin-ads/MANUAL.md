# pm connectors inspect source-linkedin-ads

```text
NAME
  pm connectors inspect source-linkedin-ads - LinkedIn Ads connector manual

SYNOPSIS
  pm connectors inspect source-linkedin-ads
  pm connectors inspect source-linkedin-ads --json
  pm credentials add <name> --connector source-linkedin-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  LinkedIn Ads catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/linkedin.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/recent-changes?view=li-lms-2024-10

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
  Changelog: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/recent-changes?view=li-lms-2024-10

CONFIGURATION
  account_ids (array): Specify the account IDs to pull data from, separated by a space. Leave this field empty if you want to pull the data from all accounts accessible by the authenticated user. See ...
  ad_analytics_reports (array)
  credentials (object)
  lookback_window (integer): How far into the past to look for records. (in days)
  num_workers (integer): The number of workers to use for the connector. This is used to limit the number of concurrent requests to the LinkedIn Ads API. If not set, the default is 3 workers.
  start_date (string) required: UTC date in the format YYYY-MM-DD. Any data before this date will not be replicated.
  secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-linkedin-ads

  # Inspect as JSON
  pm connectors inspect source-linkedin-ads --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Changelog: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/recent-changes?view=li-lms-2024-10

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
