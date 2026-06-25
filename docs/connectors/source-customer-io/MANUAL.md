# pm connectors inspect source-customer-io

```text
NAME
  pm connectors inspect source-customer-io - Customer.io connector manual

SYNOPSIS
  pm connectors inspect source-customer-io
  pm connectors inspect source-customer-io --json
  pm credentials add <name> --connector source-customer-io [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Customer.io catalog connector for https://docs.airbyte.com/integrations/sources/customer-io. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-customer-io:0.4.5 (metadata only; not executed)

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
  Customer.io API reference: https://customer.io/docs/api/
  Customer.io authentication: https://customer.io/docs/api/#section/Authentication
  Customer.io rate limits: https://customer.io/docs/api/#section/Rate-Limiting
  Customer.io Status: https://status.customer.io/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/customer-io

CONFIGURATION
  app_api_key (string) required secret
  region (string): The region of your Customer.io workspace. Select "EU" if your account is hosted in the EU region (api-eu.customer.io); otherwise leave as "US".
  start_date (string): UTC date and time in the format YYYY-MM-DDTHH:MM:SSZ. Records with an `updated` timestamp before this date will be filtered out client-side. Leave blank to sync all records.
  secret fields: app_api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/customer-io

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-customer-io

  # Inspect as JSON
  pm connectors inspect source-customer-io --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Customer.io documentation: https://docs.airbyte.com/integrations/sources/customer-io

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
