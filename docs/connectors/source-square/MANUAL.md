# pm connectors inspect source-square

```text
NAME
  pm connectors inspect source-square - Square connector manual

SYNOPSIS
  pm connectors inspect source-square
  pm connectors inspect source-square --json
  pm credentials add <name> --connector source-square [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Square catalog connector for https://docs.airbyte.com/integrations/sources/square. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-square:1.7.20 (metadata only; not executed)

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
  Square API reference: https://developer.squareup.com/reference/square
  Square authentication: https://developer.squareup.com/docs/build-basics/access-tokens
  Square API Release Notes: https://developer.squareup.com/docs/release-notes
  Square API changelog: https://developer.squareup.com/docs/changelog
  Square rate limits: https://developer.squareup.com/docs/build-basics/api-rate-limits
  Square Status: https://www.issquareup.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/square

CONFIGURATION
  credentials (object): Choose how to authenticate to Square.
  include_deleted_objects (boolean): In some streams there is an option to include deleted objects (Items, Categories, Discounts, Taxes)
  is_sandbox (boolean) required: Determines whether to use the sandbox or production environment.
  start_date (string): UTC date in the format YYYY-MM-DD. Any data before this date will not be replicated. If not set, all data will be replicated.
  secret fields: credentials.api_key, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/square

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-square

  # Inspect as JSON
  pm connectors inspect source-square --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Square documentation: https://docs.airbyte.com/integrations/sources/square

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
