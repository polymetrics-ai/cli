# pm connectors inspect source-reddit

```text
NAME
  pm connectors inspect source-reddit - Reddit connector manual

SYNOPSIS
  pm connectors inspect source-reddit
  pm connectors inspect source-reddit --json
  pm credentials add <name> --connector source-reddit [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reddit catalog connector for https://docs.airbyte.com/integrations/sources/reddit. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-reddit:0.0.54 (metadata only; not executed)

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
  Reddit API documentation: https://www.reddit.com/dev/api/
  Reddit OAuth guide: https://github.com/reddit-archive/reddit/wiki/OAuth2
  Reddit rate limits: https://github.com/reddit-archive/reddit/wiki/API#rules
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/reddit

CONFIGURATION
  api_key (string) required secret
  exact (boolean): Specifies exact keyword and reduces distractions
  include_over_18 (boolean): Includes mature content
  limit (number): Max records per page limit
  query (string): Specifies the query for searching in reddits and subreddits
  start_date (string) required
  subreddits (array): Subreddits for exploration
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/reddit

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-reddit

  # Inspect as JSON
  pm connectors inspect source-reddit --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Reddit documentation: https://docs.airbyte.com/integrations/sources/reddit

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
