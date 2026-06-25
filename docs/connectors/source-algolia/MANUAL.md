# pm connectors inspect source-algolia

```text
NAME
  pm connectors inspect source-algolia - Algolia connector manual

SYNOPSIS
  pm connectors inspect source-algolia
  pm connectors inspect source-algolia --json
  pm credentials add <name> --connector source-algolia [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Algolia catalog connector for https://docs.airbyte.com/integrations/sources/algolia. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-algolia:0.0.42 (metadata only; not executed)

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
  API reference: https://www.algolia.com/doc/rest-api/search/
  Authentication: https://www.algolia.com/doc/guides/security/api-keys/
  Rate limits: https://www.algolia.com/doc/guides/scaling/servers-clusters/#rate-limits
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/algolia

CONFIGURATION
  api_key (string) required secret
  application_id (string) required: The application ID for your application found in settings
  object_id (string): Object ID within index for search queries
  search_query (string): Search query to be used with indexes_query stream with format defined in `https://www.algolia.com/doc/rest-api/search/#tag/Search/operation/searchSingleIndex`
  start_date (string) required
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/algolia

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-algolia

  # Inspect as JSON
  pm connectors inspect source-algolia --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Algolia documentation: https://docs.airbyte.com/integrations/sources/algolia

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
