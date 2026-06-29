# pm connectors inspect source-sentry

```text
NAME
  pm connectors inspect source-sentry - Sentry connector manual

SYNOPSIS
  pm connectors inspect source-sentry
  pm connectors inspect source-sentry --json
  pm credentials add <name> --connector source-sentry [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Sentry catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/sentry.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.sentry.io/api/

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
  API Reference: https://docs.sentry.io/api/
  Changelog: https://sentry.io/changelog/
  Sentry API OpenAPI specification: https://github.com/getsentry/sentry-api-schema

CONFIGURATION
  auth_token (string) required secret: Log into Sentry and then <a href="https://sentry.io/settings/account/api/auth-tokens/">create authentication tokens</a>.For self-hosted, you can find or create authentication to...
  discover_fields (array): Fields to retrieve when fetching discover events
  hostname (string): Host name of Sentry API server.For self-hosted, specify your host name here. Otherwise, leave it empty.
  num_workers (integer): The number of worker threads to use for the sync. The default of 5 is safe for all Sentry API plans. Testing showed concurrency of 7 is optimal; increase if your Sentry plan sup...
  organization (string) required: The slug of the organization the groups belong to.
  project (string) required: The name (slug) of the Project you want to sync.
  secret fields: auth_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-sentry

  # Inspect as JSON
  pm connectors inspect source-sentry --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  API Reference: https://docs.sentry.io/api/
  Changelog: https://sentry.io/changelog/
  Sentry API OpenAPI specification: https://github.com/getsentry/sentry-api-schema

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
