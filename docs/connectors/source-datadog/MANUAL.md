# pm connectors inspect source-datadog

```text
NAME
  pm connectors inspect source-datadog - Datadog connector manual

SYNOPSIS
  pm connectors inspect source-datadog
  pm connectors inspect source-datadog --json
  pm credentials add <name> --connector source-datadog [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Datadog catalog connector for https://docs.airbyte.com/integrations/sources/datadog. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-datadog:2.0.24 (metadata only; not executed)

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
  Datadog API reference: https://docs.datadoghq.com/api/latest/
  Datadog authentication: https://docs.datadoghq.com/account_management/api-app-keys/
  Datadog rate limits: https://docs.datadoghq.com/api/latest/rate-limits/
  Datadog Status: https://status.datadoghq.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/datadog

CONFIGURATION
  api_key (string) required secret: Datadog API key
  application_key (string) required secret: Datadog application key
  end_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Data after this date will not be replicated. An empty value will represent the current datetime for each execution. This ju...
  max_records_per_request (integer): Maximum number of records to collect per request.
  queries (array): List of queries to be run and used as inputs.
  query (string): The search query. This just applies to Incremental syncs. If empty, it'll collect all logs.
  site (string): The site where Datadog data resides in.
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated. This just applies to Incremental syncs.
  secret fields: api_key, application_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/datadog

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-datadog

  # Inspect as JSON
  pm connectors inspect source-datadog --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Datadog documentation: https://docs.airbyte.com/integrations/sources/datadog

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
