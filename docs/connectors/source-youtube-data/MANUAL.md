# pm connectors inspect source-youtube-data

```text
NAME
  pm connectors inspect source-youtube-data - Youtube Data connector manual

SYNOPSIS
  pm connectors inspect source-youtube-data
  pm connectors inspect source-youtube-data --json
  pm credentials add <name> --connector source-youtube-data [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Youtube Data catalog connector for https://docs.airbyte.com/integrations/sources/youtube-data. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-youtube-data:0.0.58 (metadata only; not executed)

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
  YouTube Data API v3: https://developers.google.com/youtube/v3
  Google OAuth 2.0: https://developers.google.com/identity/protocols/oauth2
  YouTube API changelog: https://developers.google.com/youtube/v3/revision_history
  YouTube API quotas: https://developers.google.com/youtube/v3/getting-started#quota
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/youtube-data

CONFIGURATION
  channel_ids (array) required
  credentials (object) required: Authentication method
  secret fields: credentials.api_key, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/youtube-data

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-youtube-data

  # Inspect as JSON
  pm connectors inspect source-youtube-data --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Youtube Data documentation: https://docs.airbyte.com/integrations/sources/youtube-data

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
