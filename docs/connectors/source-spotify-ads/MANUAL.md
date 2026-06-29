# pm connectors inspect source-spotify-ads

```text
NAME
  pm connectors inspect source-spotify-ads - Spotify Ads connector manual

SYNOPSIS
  pm connectors inspect source-spotify-ads
  pm connectors inspect source-spotify-ads --json
  pm credentials add <name> --connector source-spotify-ads [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Spotify Ads catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

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
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Spotify Ads API: https://ads.spotify.com/en-US/api/

CONFIGURATION
  ad_account_id (string) required: The ID of the Spotify Ad Account you want to sync data from.
  client_id (string) required secret: The Client ID of your Spotify Developer application.
  client_secret (string) required secret: The Client Secret of your Spotify Developer application.
  fields (array) required: List of fields to include in the campaign performance report. Choose from available metrics.
  refresh_token (string) required secret: The Refresh Token obtained from the initial OAuth 2.0 authorization flow.
  start_date (string) required: The date to start syncing data from, in YYYY-MM-DD format.
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
  pm connectors inspect source-spotify-ads

  # Inspect as JSON
  pm connectors inspect source-spotify-ads --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Spotify Ads API: https://ads.spotify.com/en-US/api/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
