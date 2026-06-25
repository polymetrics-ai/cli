# pm connectors inspect source-openweather

```text
NAME
  pm connectors inspect source-openweather - Openweather connector manual

SYNOPSIS
  pm connectors inspect source-openweather
  pm connectors inspect source-openweather --json
  pm credentials add <name> --connector source-openweather [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Openweather catalog connector for https://docs.airbyte.com/integrations/sources/openweather. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-openweather:0.3.54 (metadata only; not executed)

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
  OpenWeather API documentation: https://openweathermap.org/api
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/openweather

CONFIGURATION
  appid (string) required secret: API KEY
  lang (string): You can use lang parameter to get the output in your language. The contents of the description field will be translated. See <a href="https://openweathermap.org/api/one-call-api...
  lat (string) required: Latitude, decimal (-90; 90). If you need the geocoder to automatic convert city names and zip-codes to geo coordinates and the other way around, please use the OpenWeather Geoco...
  lon (string) required: Longitude, decimal (-180; 180). If you need the geocoder to automatic convert city names and zip-codes to geo coordinates and the other way around, please use the OpenWeather Ge...
  only_current (boolean): True for particular day
  units (string): Units of measurement. standard, metric and imperial units are available. If you do not use the units parameter, standard units will be applied by default.
  secret fields: appid

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/openweather

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-openweather

  # Inspect as JSON
  pm connectors inspect source-openweather --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Openweather documentation: https://docs.airbyte.com/integrations/sources/openweather

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
