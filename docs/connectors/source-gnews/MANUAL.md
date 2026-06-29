# pm connectors inspect source-gnews

```text
NAME
  pm connectors inspect source-gnews - GNews connector manual

SYNOPSIS
  pm connectors inspect source-gnews
  pm connectors inspect source-gnews --json
  pm credentials add <name> --connector source-gnews [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  GNews catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/gnews.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://gnews.io/docs/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

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
  GNews API documentation: https://gnews.io/docs/

CONFIGURATION
  api_key (string) required secret: API Key
  country (string): This parameter allows you to specify the country where the news articles returned by the API were published, the contents of the articles are not necessarily related to the spec...
  end_date (string): This parameter allows you to filter the articles that have a publication date smaller than or equal to the specified value. The date must respect the following format: YYYY-MM-D...
  in (array): This parameter allows you to choose in which attributes the keywords are searched. The attributes that can be set are title, description and content. It is possible to combine s...
  language (string)
  nullable (array): This parameter allows you to specify the attributes that you allow to return null values. The attributes that can be set are title, description and content. It is possible to co...
  query (string) required: This parameter allows you to specify your search keywords to find the news articles you are looking for. The keywords will be used to return the most relevant articles. It is po...
  sortby (string): This parameter allows you to choose with which type of sorting the articles should be returned. Two values are possible: - publishedAt = sort by publication date, the articles w...
  start_date (string): This parameter allows you to filter the articles that have a publication date greater than or equal to the specified value. The date must respect the following format: YYYY-MM-D...
  top_headlines_query (string): This parameter allows you to specify your search keywords to find the news articles you are looking for. The keywords will be used to return the most relevant articles. It is po...
  top_headlines_topic (string): This parameter allows you to change the category for the request.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-gnews

  # Inspect as JSON
  pm connectors inspect source-gnews --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  GNews API documentation: https://gnews.io/docs/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
