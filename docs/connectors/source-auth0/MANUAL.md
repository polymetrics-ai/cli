# pm connectors inspect source-auth0

```text
NAME
  pm connectors inspect source-auth0 - Auth0 connector manual

SYNOPSIS
  pm connectors inspect source-auth0
  pm connectors inspect source-auth0 --json
  pm credentials add <name> --connector source-auth0 [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Auth0 catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/auth0.svg
  source: official
  review_status: official_verified
  review_url: https://auth0.com/docs/api/management/v2

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
  Auth0 documentation: https://auth0.com/docs/api/management/v2

CONFIGURATION
  base_url (string) required: The Authentication API is served over HTTPS. All URLs referenced in the documentation have the following base `https://YOUR_DOMAIN`
  credentials (object) required
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  secret fields: credentials.access_token, credentials.client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-auth0

  # Inspect as JSON
  pm connectors inspect source-auth0 --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Auth0 documentation: https://auth0.com/docs/api/management/v2

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
