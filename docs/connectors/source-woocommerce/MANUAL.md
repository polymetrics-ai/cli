# pm connectors inspect source-woocommerce

```text
NAME
  pm connectors inspect source-woocommerce - WooCommerce connector manual

SYNOPSIS
  pm connectors inspect source-woocommerce
  pm connectors inspect source-woocommerce --json
  pm credentials add <name> --connector source-woocommerce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  WooCommerce catalog connector for https://docs.airbyte.com/integrations/sources/woocommerce. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-woocommerce:0.5.41 (metadata only; not executed)

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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  WooCommerce REST API: https://woocommerce.github.io/woocommerce-rest-api-docs/
  WooCommerce authentication: https://woocommerce.github.io/woocommerce-rest-api-docs/#authentication
  WooCommerce Developer Changelog: https://developer.woocommerce.com/changelog/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/woocommerce

CONFIGURATION
  api_key (string) required secret: Customer Key for API in WooCommerce shop
  api_secret (string) required secret: Customer Secret for API in WooCommerce shop
  num_workers (integer): The number of worker threads to use for the sync. Higher values can speed up syncs but may hit rate limits. WooCommerce API rate limits depend on the hosting provider. More info...
  shop (string) required: The name of the store. For https://EXAMPLE.com, the shop name is 'EXAMPLE.com'.
  start_date (string) required: The date you would like to replicate data from. Format: YYYY-MM-DD
  secret fields: api_key, api_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/woocommerce

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-woocommerce

  # Inspect as JSON
  pm connectors inspect source-woocommerce --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  WooCommerce documentation: https://docs.airbyte.com/integrations/sources/woocommerce

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
