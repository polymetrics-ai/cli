# pm connectors inspect source-amazon-ads

```text
NAME
  pm connectors inspect source-amazon-ads - Amazon Ads connector manual

SYNOPSIS
  pm connectors inspect source-amazon-ads
  pm connectors inspect source-amazon-ads --json
  pm credentials add <name> --connector source-amazon-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Amazon Ads catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/amazonads.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://advertising.amazon.com/API/docs/en-us/release-notes/deprecations

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
  Advertising API: https://advertising.amazon.com/API/docs/en-us
  Authorization: https://advertising.amazon.com/API/docs/en-us/get-started/authorization
  All releases: https://advertising.amazon.com/API/docs/en-us/release-notes/index
  Deprecations: https://advertising.amazon.com/API/docs/en-us/release-notes/deprecations
  Rate limits: https://advertising.amazon.com/API/docs/en-us/get-started/developer-notes#rate-limiting

CONFIGURATION
  auth_type (string)
  client_id (string) required secret: The client ID of your Amazon Ads developer application. See the <a href="https://advertising.amazon.com/API/docs/en-us/get-started/generate-api-tokens#retrieve-your-client-id-an...
  client_secret (string) required secret: The client secret of your Amazon Ads developer application. See the <a href="https://advertising.amazon.com/API/docs/en-us/get-started/generate-api-tokens#retrieve-your-client-i...
  look_back_window (integer): The amount of days to go back in time to get the updated data from Amazon Ads
  marketplace_ids (array): Marketplace IDs you want to fetch data for. Note: If Profile IDs are also selected, profiles will be selected if they match the Profile ID OR the Marketplace ID.
  num_workers (integer): The number of worker threads to use for the sync.
  profiles (array): Profile IDs you want to fetch data for. The Amazon Ads source connector supports only profiles with seller and vendor type, profiles with agency type will be ignored. See <a hre...
  refresh_token (string) required secret: Amazon Ads refresh token. See the <a href="https://advertising.amazon.com/API/docs/en-us/get-started/generate-api-tokens">docs</a> for more information on how to obtain this token.
  region (string): Region to pull data from (EU/NA/FE). See <a href="https://advertising.amazon.com/API/docs/en-us/info/api-overview#api-endpoints">docs</a> for more details.
  start_date (string): The Start date for collecting reports, should not be more than 60 days in the past. In YYYY-MM-DD format
  secret fields: client_id, client_secret, refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-amazon-ads

  # Inspect as JSON
  pm connectors inspect source-amazon-ads --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Advertising API: https://advertising.amazon.com/API/docs/en-us
  Authorization: https://advertising.amazon.com/API/docs/en-us/get-started/authorization
  All releases: https://advertising.amazon.com/API/docs/en-us/release-notes/index
  Deprecations: https://advertising.amazon.com/API/docs/en-us/release-notes/deprecations
  Rate limits: https://advertising.amazon.com/API/docs/en-us/get-started/developer-notes#rate-limiting

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
