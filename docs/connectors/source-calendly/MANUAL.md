# pm connectors inspect source-calendly

```text
NAME
  pm connectors inspect source-calendly - Calendly connector manual

SYNOPSIS
  pm connectors inspect source-calendly
  pm connectors inspect source-calendly --json
  pm credentials add <name> --connector source-calendly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Calendly catalog connector for https://docs.airbyte.com/integrations/sources/calendly. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-calendly:0.1.42 (metadata only; not executed)

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
  Calendly API reference: https://developer.calendly.com/api-docs
  Calendly authentication: https://developer.calendly.com/getting-started
  Calendly API rate limits: https://developer.calendly.com/api-docs/ZG9jOjM2MzE2MDM4-api-conventions#rate-limiting
  Calendly Status: https://status.calendly.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/calendly

CONFIGURATION
  api_key (string) required secret: Go to Integrations → API & Webhooks to obtain your bearer token. https://calendly.com/integrations/api_webhooks
  lookback_days (number): Number of days to be subtracted from the last cutoff date before starting to sync the `scheduled_events` stream.
  start_date (string) required
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/calendly

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-calendly

  # Inspect as JSON
  pm connectors inspect source-calendly --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Calendly documentation: https://docs.airbyte.com/integrations/sources/calendly

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
