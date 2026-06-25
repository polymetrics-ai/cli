# pm connectors inspect destination-customer-io

```text
NAME
  pm connectors inspect destination-customer-io - Customer IO connector manual

SYNOPSIS
  pm connectors inspect destination-customer-io
  pm connectors inspect destination-customer-io --json
  pm credentials add <name> --connector destination-customer-io [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Customer IO catalog connector for https://docs.airbyte.com/integrations/destinations/customer-io. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-customer-io:0.0.11 (metadata only; not executed)

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
  family: destination_writer
  priority_wave: 3
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  Customer.io API documentation: https://customer.io/docs/api/
  Authentication: https://customer.io/docs/api/app/#tag/Authentication
  Rate limits: https://customer.io/docs/api/app/#tag/Rate-Limits
  Customer.io Status: https://status.customer.io/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/customer-io

CONFIGURATION
  credentials (object) required: Enter the site ID and API key to authenticate.
  object_storage_config (object)
  secret fields: credentials.apiKey, credentials.siteId, object_storage_config.access_key_id, object_storage_config.secret_access_key

SYNC MODES
  supported sync modes: append
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/customer-io

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-customer-io

  # Inspect as JSON
  pm connectors inspect destination-customer-io --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Customer IO documentation: https://docs.airbyte.com/integrations/destinations/customer-io

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
