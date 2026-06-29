# pm connectors inspect source-strava

```text
NAME
  pm connectors inspect source-strava - Strava connector manual

SYNOPSIS
  pm connectors inspect source-strava
  pm connectors inspect source-strava --json
  pm credentials add <name> --connector source-strava [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Strava catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/strava.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.strava.com/docs/reference/

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
  Strava API reference: https://developers.strava.com/docs/reference/
  Strava authentication: https://developers.strava.com/docs/authentication/
  Strava rate limits: https://developers.strava.com/docs/rate-limits/

CONFIGURATION
  athlete_id (integer) required: The Athlete ID of your Strava developer application.
  auth_type (string)
  client_id (string) required: The Client ID of your Strava developer application.
  client_secret (string) required secret: The Client Secret of your Strava developer application.
  refresh_token (string) required secret: The Refresh Token with the activity: read_all permissions.
  start_date (string) required: UTC date and time. Any data before this date will not be replicated.
  secret fields: client_secret, refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-strava

  # Inspect as JSON
  pm connectors inspect source-strava --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Strava API reference: https://developers.strava.com/docs/reference/
  Strava authentication: https://developers.strava.com/docs/authentication/
  Strava rate limits: https://developers.strava.com/docs/rate-limits/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
