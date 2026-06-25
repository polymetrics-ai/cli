# pm connectors inspect source-pardot

```text
NAME
  pm connectors inspect source-pardot - Pardot connector manual

SYNOPSIS
  pm connectors inspect source-pardot
  pm connectors inspect source-pardot --json
  pm credentials add <name> --connector source-pardot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Pardot catalog connector for https://docs.airbyte.com/integrations/sources/pardot. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-pardot:1.0.46 (metadata only; not executed)

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
  Pardot API reference: https://developer.salesforce.com/docs/marketing/pardot/overview
  Pardot authentication: https://developer.salesforce.com/docs/marketing/pardot/guide/authentication.html
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/pardot

CONFIGURATION
  client_id (string) required secret: The Consumer Key that can be found when viewing your app in Salesforce
  client_secret (string) required secret: The Consumer Secret that can be found when viewing your app in Salesforce
  is_sandbox (boolean): Whether or not the the app is in a Salesforce sandbox. If you do not know what this, assume it is false.
  pardot_business_unit_id (string) required: Pardot Business ID, can be found at Setup > Pardot > Pardot Account Setup
  refresh_token (string) required secret: Salesforce Refresh Token used for Airbyte to access your Salesforce account. If you don't know what this is, follow this <a href="https://medium.com/@bpmmendis94/obtain-access-r...
  start_date (string): UTC date and time in the format 2000-01-01T00:00:00Z. Any data before this date will not be replicated. Defaults to the year Pardot was released.
  v5_page_size (integer): The maximum number of records to return per request
  secret fields: client_id, client_secret, refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/pardot

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-pardot

  # Inspect as JSON
  pm connectors inspect source-pardot --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Pardot documentation: https://docs.airbyte.com/integrations/sources/pardot

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
