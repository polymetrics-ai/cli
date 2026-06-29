# pm connectors inspect source-youtube-analytics

```text
NAME
  pm connectors inspect source-youtube-analytics - YouTube Analytics connector manual

SYNOPSIS
  pm connectors inspect source-youtube-analytics
  pm connectors inspect source-youtube-analytics --json
  pm credentials add <name> --connector source-youtube-analytics [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  YouTube Analytics catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/youtube-analytics.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/youtube/analytics

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  YouTube Analytics API: https://developers.google.com/youtube/analytics
  Google OAuth 2.0: https://developers.google.com/identity/protocols/oauth2
  YouTube API changelog: https://developers.google.com/youtube/v3/revision_history
  YouTube API quotas: https://developers.google.com/youtube/v3/getting-started#quota

CONFIGURATION
  content_owner_id (string): The ID of the content owner for whom the API request is being made. This is useful if you manage multiple YouTube channels and need to specify which content owner's data to retr...
  credentials (object) required
  secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-youtube-analytics

  # Inspect as JSON
  pm connectors inspect source-youtube-analytics --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  YouTube Analytics API: https://developers.google.com/youtube/analytics
  Google OAuth 2.0: https://developers.google.com/identity/protocols/oauth2
  YouTube API changelog: https://developers.google.com/youtube/v3/revision_history
  YouTube API quotas: https://developers.google.com/youtube/v3/getting-started#quota

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
