# pm connectors inspect source-newsdata-io

```text
NAME
  pm connectors inspect source-newsdata-io - NewsData.io connector manual

SYNOPSIS
  pm connectors inspect source-newsdata-io
  pm connectors inspect source-newsdata-io --json
  pm credentials add <name> --connector source-newsdata-io [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  NewsData.io catalog connector for https://docs.airbyte.com/integrations/sources/newsdata-io. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-newsdata-io:0.0.53 (metadata only; not executed)

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
  NewsData.io API documentation: https://newsdata.io/documentation
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/newsdata-io

CONFIGURATION
  api_key (string) required secret
  categories (array): Search the news articles for a specific category. You can add up to 5 categories in a single query.
  countries (array): Search the news articles from a specific country. You can add up to 5 countries in a single query. Example: au, jp, br
  domains (array): Search the news articles for specific domains or news sources. You can add up to 5 domains in a single query.
  end_date (string): Choose an end date. Now UTC is default value
  languages (array): Search the news articles for a specific language. You can add up to 5 languages in a single query.
  search_query (string): Search news articles for specific keywords or phrases present in the news title, content, URL, meta keywords and meta description.
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
  https://docs.airbyte.com/integrations/sources/newsdata-io

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-newsdata-io

  # Inspect as JSON
  pm connectors inspect source-newsdata-io --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  NewsData.io documentation: https://docs.airbyte.com/integrations/sources/newsdata-io

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
