# pm connectors inspect source-prestashop

```text
NAME
  pm connectors inspect source-prestashop - PrestaShop connector manual

SYNOPSIS
  pm connectors inspect source-prestashop
  pm connectors inspect source-prestashop --json
  pm credentials add <name> --connector source-prestashop [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  PrestaShop catalog connector for https://docs.airbyte.com/integrations/sources/prestashop. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-prestashop:1.2.7 (metadata only; not executed)

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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/prestashop

CONFIGURATION
  access_key (string) required secret: Your PrestaShop access key. See <a href="https://devdocs.prestashop.com/1.7/webservice/tutorials/creating-access/#create-an-access-key"> the docs </a> for info on how to obtain ...
  start_date (string) required: The Start date in the format YYYY-MM-DD.
  url (string) required: Shop URL without trailing slash.
  secret fields: access_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/prestashop

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-prestashop

  # Inspect as JSON
  pm connectors inspect source-prestashop --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  PrestaShop documentation: https://docs.airbyte.com/integrations/sources/prestashop

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
