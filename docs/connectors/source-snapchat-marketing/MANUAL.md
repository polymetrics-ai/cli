# pm connectors inspect source-snapchat-marketing

```text
NAME
  pm connectors inspect source-snapchat-marketing - Snapchat Marketing connector manual

SYNOPSIS
  pm connectors inspect source-snapchat-marketing
  pm connectors inspect source-snapchat-marketing --json
  pm credentials add <name> --connector source-snapchat-marketing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Snapchat Marketing catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/snapchat.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.snap.com/api/marketing-api/Ads-API/announcements

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
  Ads API announcements: https://developers.snap.com/api/marketing-api/Ads-API/announcements

CONFIGURATION
  action_report_time (string): Specifies the principle for conversion reporting.
  ad_account_ids (array): Ad Account IDs of the ad accounts to retrieve
  client_id (string) required secret: The Client ID of your Snapchat developer application.
  client_secret (string) required secret: The Client Secret of your Snapchat developer application.
  end_date (string): Date in the format 2017-01-25. Any data after this date will not be replicated.
  organization_ids (array): The IDs of the organizations to retrieve
  refresh_token (string) required secret: Refresh Token to renew the expired Access Token.
  start_date (string): Date in the format 2022-01-01. Any data before this date will not be replicated.
  swipe_up_attribution_window (string): Attribution window for swipe ups.
  view_attribution_window (string): Attribution window for views.
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
  pm connectors inspect source-snapchat-marketing

  # Inspect as JSON
  pm connectors inspect source-snapchat-marketing --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Ads API announcements: https://developers.snap.com/api/marketing-api/Ads-API/announcements

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
