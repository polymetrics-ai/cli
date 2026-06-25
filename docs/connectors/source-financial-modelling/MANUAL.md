# pm connectors inspect source-financial-modelling

```text
NAME
  pm connectors inspect source-financial-modelling - Financial Modelling connector manual

SYNOPSIS
  pm connectors inspect source-financial-modelling
  pm connectors inspect source-financial-modelling --json
  pm credentials add <name> --connector source-financial-modelling [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Financial Modelling catalog connector for https://docs.airbyte.com/integrations/sources/financial-modelling. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-financial-modelling:0.0.53 (metadata only; not executed)

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
  Financial Modeling Prep API: https://site.financialmodelingprep.com/developer/docs
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/financial-modelling

CONFIGURATION
  api_key (string) required secret
  exchange (string): The stock exchange : AMEX, AMS, AQS, ASX, ATH, BER, BME, BRU, BSE, BUD, BUE, BVC, CAI, CBOE, CNQ, CPH, DFM, DOH, DUS, DXE, EGX, EURONEXT, HAM, HEL, HKSE, ICE, IOB, IST, JKT, JNB...
  marketcaplowerthan (string): Used in screener to filter out stocks with a market cap lower than the give marketcap
  marketcapmorethan (string): Used in screener to filter out stocks with a market cap more than the give marketcap
  start_date (string) required
  time_frame (string): For example 1min, 5min, 15min, 30min, 1hour, 4hour
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/financial-modelling

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-financial-modelling

  # Inspect as JSON
  pm connectors inspect source-financial-modelling --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Financial Modelling documentation: https://docs.airbyte.com/integrations/sources/financial-modelling

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
