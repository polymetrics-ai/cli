# pm connectors inspect source-gocardless

```text
NAME
  pm connectors inspect source-gocardless - GoCardless connector manual

SYNOPSIS
  pm connectors inspect source-gocardless
  pm connectors inspect source-gocardless --json
  pm credentials add <name> --connector source-gocardless [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  GoCardless catalog connector for https://docs.airbyte.com/integrations/sources/gocardless. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-gocardless:0.2.25 (metadata only; not executed)

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
  GoCardless API reference: https://developer.gocardless.com/api-reference/
  GoCardless authentication: https://developer.gocardless.com/getting-started/api/making-your-first-api-request/
  GoCardless rate limits: https://developer.gocardless.com/api-reference/#making-requests-rate-limiting
  GoCardless Status: https://status.gocardless.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/gocardless

CONFIGURATION
  access_token (string) required secret: Gocardless API TOKEN
  gocardless_environment (string) required: Environment you are trying to connect to.
  gocardless_version (string) required: GoCardless version. This is a date. You can find the latest here: https://developer.gocardless.com/api-reference/#api-usage-making-requests
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  secret fields: access_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/gocardless

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-gocardless

  # Inspect as JSON
  pm connectors inspect source-gocardless --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  GoCardless documentation: https://docs.airbyte.com/integrations/sources/gocardless

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
