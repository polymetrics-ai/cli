# pm connectors inspect source-wikipedia-pageviews

```text
NAME
  pm connectors inspect source-wikipedia-pageviews - Wikipedia Pageviews connector manual

SYNOPSIS
  pm connectors inspect source-wikipedia-pageviews
  pm connectors inspect source-wikipedia-pageviews --json
  pm credentials add <name> --connector source-wikipedia-pageviews [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Wikipedia Pageviews catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/wikipedia-pageviews.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://wikitech.wikimedia.org/wiki/Analytics/AQS/Pageviews

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
  Wikimedia Pageviews API: https://wikitech.wikimedia.org/wiki/Analytics/AQS/Pageviews

CONFIGURATION
  access (string) required: If you want to filter by access method, use one of desktop, mobile-app or mobile-web. If you are interested in pageviews regardless of access method, use all-access.
  agent (string) required: If you want to filter by agent type, use one of user, automated or spider. If you are interested in pageviews regardless of agent type, use all-agents.
  article (string) required: The title of any article in the specified project. Any spaces should be replaced with underscores. It also should be URI-encoded, so that non-URI-safe characters like %, / or ? ...
  country (string) required: The ISO 3166-1 alpha-2 code of a country for which to retrieve top articles.
  end (string) required: The date of the last day to include, in YYYYMMDD or YYYYMMDDHH format.
  project (string) required: If you want to filter by project, use the domain of any Wikimedia project.
  start (string) required: The date of the first day to include, in YYYYMMDD or YYYYMMDDHH format. Also serves as the date to retrieve data for the top articles.

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-wikipedia-pageviews

  # Inspect as JSON
  pm connectors inspect source-wikipedia-pageviews --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Wikimedia Pageviews API: https://wikitech.wikimedia.org/wiki/Analytics/AQS/Pageviews

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
