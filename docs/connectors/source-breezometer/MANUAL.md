# pm connectors inspect source-breezometer

```text
NAME
  pm connectors inspect source-breezometer - Breezometer connector manual

SYNOPSIS
  pm connectors inspect source-breezometer
  pm connectors inspect source-breezometer --json
  pm credentials add <name> --connector source-breezometer [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Breezometer catalog connector for https://docs.airbyte.com/integrations/sources/breezometer. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-breezometer:0.2.24 (metadata only; not executed)

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
  BreezoMeter API documentation: https://docs.breezometer.com/api-documentation/
  BreezoMeter authentication: https://docs.breezometer.com/api-documentation/introduction/#authentication
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/breezometer

CONFIGURATION
  api_key (string) required secret: Your API Access Key. See <a href="https://docs.breezometer.com/api-documentation/introduction/#authentication/">here</a>.
  days_to_forecast (integer): Number of days to forecast. Minimum 1, maximum 3. Valid for Polen and Weather Forecast streams.
  historic_hours (integer): Number of hours retireve from Air Quality History stream. Minimum 1, maximum 720.
  hours_to_forecast (integer): Number of hours to forecast. Minimum 1, maximum 96. Valid for Air Quality Forecast stream.
  latitude (string) required: Latitude of the monitored location.
  longitude (string) required: Longitude of the monitored location.
  radius (integer): Desired radius from the location provided. Minimum 5, maximum 100. Valid for Wildfires streams.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/breezometer

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-breezometer

  # Inspect as JSON
  pm connectors inspect source-breezometer --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Breezometer documentation: https://docs.airbyte.com/integrations/sources/breezometer

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
