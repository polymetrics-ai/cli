# pm connectors inspect source-finage

```text
NAME
  pm connectors inspect source-finage - Finage connector manual

SYNOPSIS
  pm connectors inspect source-finage
  pm connectors inspect source-finage --json
  pm credentials add <name> --connector source-finage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Finage catalog connector for https://docs.airbyte.com/integrations/sources/finage. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-finage:0.0.55 (metadata only; not executed)

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
  Finage API documentation: https://finage.co.uk/docs/api
  Finage authentication: https://finage.co.uk/docs/api/getting-started
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/finage

CONFIGURATION
  api_key (string) required secret
  period (string): Time period. Default is 10
  start_date (string) required
  symbols (array) required: List of symbols
  tech_indicator_type (string): One of DEMA, EMA, SMA, WMA, RSI, TEMA, Williams, ADX
  time (string)
  time_aggregates (string): Size of the time
  time_period (string): Time Period for cash flow stmts
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/finage

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-finage

  # Inspect as JSON
  pm connectors inspect source-finage --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Finage documentation: https://docs.airbyte.com/integrations/sources/finage

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
