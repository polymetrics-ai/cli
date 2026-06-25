# pm connectors inspect source-plausible

```text
NAME
  pm connectors inspect source-plausible - Plausible connector manual

SYNOPSIS
  pm connectors inspect source-plausible
  pm connectors inspect source-plausible --json
  pm credentials add <name> --connector source-plausible [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Plausible catalog connector for https://docs.airbyte.com/integrations/sources/plausible. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-plausible:0.2.15 (metadata only; not executed)

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
  Plausible Analytics API: https://plausible.io/docs/stats-api
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/plausible

CONFIGURATION
  api_key (string) required secret: Plausible API Key. See the <a href="https://plausible.io/docs/stats-api">docs</a> for information on how to generate this key.
  api_url (string): The API URL of your plausible instance. Change this if you self-host plausible. The default is https://plausible.io/api/v1/stats
  site_id (string) required: The domain of the site you want to retrieve data for. Enter the name of your site as configured on Plausible, i.e., excluding "https://" and "www". Can be retrieved from the 'do...
  start_date (string): Start date for data to retrieve, in ISO-8601 format.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/plausible

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-plausible

  # Inspect as JSON
  pm connectors inspect source-plausible --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Plausible documentation: https://docs.airbyte.com/integrations/sources/plausible

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
