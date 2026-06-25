# pm connectors inspect source-coingecko-coins

```text
NAME
  pm connectors inspect source-coingecko-coins - CoinGecko Coins connector manual

SYNOPSIS
  pm connectors inspect source-coingecko-coins
  pm connectors inspect source-coingecko-coins --json
  pm credentials add <name> --connector source-coingecko-coins [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  CoinGecko Coins catalog connector for https://docs.airbyte.com/integrations/sources/coingecko-coins. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-coingecko-coins:0.2.26 (metadata only; not executed)

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
  CoinGecko API documentation: https://www.coingecko.com/en/api/documentation
  CoinGecko rate limits: https://www.coingecko.com/en/api/pricing
  CoinGecko Status: https://status.coingecko.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/coingecko-coins

CONFIGURATION
  api_key (string) secret: API Key (for pro users)
  coin_id (string) required: CoinGecko coin ID (e.g. bitcoin). Can be retrieved from the `/coins/list` endpoint.
  days (string) required: The number of days of data for market chart.
  end_date (string): The end date for the historical data stream in dd-mm-yyyy format.
  start_date (string) required: The start date for the historical data stream in dd-mm-yyyy format.
  vs_currency (string) required: The target currency of market data (e.g. usd, eur, jpy, etc.)
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/coingecko-coins

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-coingecko-coins

  # Inspect as JSON
  pm connectors inspect source-coingecko-coins --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  CoinGecko Coins documentation: https://docs.airbyte.com/integrations/sources/coingecko-coins

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
