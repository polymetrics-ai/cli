# pm connectors inspect source-awin-advertiser

```text
NAME
  pm connectors inspect source-awin-advertiser - AWIN Advertiser connector manual

SYNOPSIS
  pm connectors inspect source-awin-advertiser
  pm connectors inspect source-awin-advertiser --json
  pm credentials add <name> --connector source-awin-advertiser [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  AWIN Advertiser catalog connector for https://docs.airbyte.com/integrations/sources/awin-advertiser. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-awin-advertiser:0.0.27 (metadata only; not executed)

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
  API documentation: https://wiki.awin.com/index.php/Advertiser_API
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/awin-advertiser

CONFIGURATION
  advertiserId (string) required: Your Awin Advertiser ID. You can find this in your Awin dashboard or account settings.
  api_key (string) required secret: Your Awin API key. Generate this from your Awin account under API Credentials.
  lookback_days (integer) required: Number of days to look back on each sync to catch any updates to existing records.
  start_date (string) required: Start date for data replication in YYYY-MM-DD format
  step_increment (string) required: The time window size for each API request in ISO8601 duration format. For the campaign performance stream, Awin API explicitly limits the period between startDate and endDate to...
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/awin-advertiser

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-awin-advertiser

  # Inspect as JSON
  pm connectors inspect source-awin-advertiser --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  AWIN Advertiser documentation: https://docs.airbyte.com/integrations/sources/awin-advertiser

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
