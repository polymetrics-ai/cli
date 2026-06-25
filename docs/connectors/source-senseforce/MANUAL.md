# pm connectors inspect source-senseforce

```text
NAME
  pm connectors inspect source-senseforce - Senseforce connector manual

SYNOPSIS
  pm connectors inspect source-senseforce
  pm connectors inspect source-senseforce --json
  pm credentials add <name> --connector source-senseforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Senseforce catalog connector for https://docs.airbyte.com/integrations/sources/senseforce. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-senseforce:0.2.27 (metadata only; not executed)

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
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/senseforce

CONFIGURATION
  access_token (string) required secret: Your API access token. See <a href="https://manual.senseforce.io/manual/sf-platform/public-api/get-your-access-token/">here</a>. The toke is case sensitive.
  backend_url (string) required: Your Senseforce API backend URL. This is the URL shown during the Login screen. See <a href="https://manual.senseforce.io/manual/sf-platform/public-api/get-your-access-token/">h...
  dataset_id (string) required: The ID of the dataset you want to synchronize. The ID can be found in the URL when opening the dataset. See <a href="https://manual.senseforce.io/manual/sf-platform/public-api/g...
  slice_range (integer): The time increment used by the connector when requesting data from the Senseforce API. The bigger the value is, the less requests will be made and faster the sync will be. On th...
  start_date (string) required: UTC date and time in the format 2017-01-25. Only data with "Timestamp" after this date will be replicated. Important note: This start date must be set to the first day of where ...
  secret fields: access_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/senseforce

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-senseforce

  # Inspect as JSON
  pm connectors inspect source-senseforce --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Senseforce documentation: https://docs.airbyte.com/integrations/sources/senseforce

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
