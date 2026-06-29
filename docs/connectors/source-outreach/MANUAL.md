# pm connectors inspect source-outreach

```text
NAME
  pm connectors inspect source-outreach - Outreach connector manual

SYNOPSIS
  pm connectors inspect source-outreach
  pm connectors inspect source-outreach --json
  pm credentials add <name> --connector source-outreach [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Outreach catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/outreach.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.outreach.io/api/v2/docs

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
  Outreach API reference: https://api.outreach.io/api/v2/docs
  Outreach authentication: https://api.outreach.io/api/v2/docs#authentication
  Outreach rate limits: https://api.outreach.io/api/v2/docs#rate-limiting

CONFIGURATION
  client_id (string) required: The Client ID of your Outreach developer application.
  client_secret (string) required secret: The Client Secret of your Outreach developer application.
  redirect_uri (string) required: A Redirect URI is the location where the authorization server sends the user once the app has been successfully authorized and granted an authorization code or access token.
  refresh_token (string) required secret: The token for obtaining the new access token.
  start_date (string) required: The date from which you'd like to replicate data for Outreach API, in the format YYYY-MM-DDT00:00:00.000Z. All data generated after this date will be replicated.
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
  pm connectors inspect source-outreach

  # Inspect as JSON
  pm connectors inspect source-outreach --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Outreach API reference: https://api.outreach.io/api/v2/docs
  Outreach authentication: https://api.outreach.io/api/v2/docs#authentication
  Outreach rate limits: https://api.outreach.io/api/v2/docs#rate-limiting

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
