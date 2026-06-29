# pm connectors inspect source-pocket

```text
NAME
  pm connectors inspect source-pocket - Pocket connector manual

SYNOPSIS
  pm connectors inspect source-pocket
  pm connectors inspect source-pocket --json
  pm credentials add <name> --connector source-pocket [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Pocket catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pocket.svg
  source: upstream_registry
  review_status: upstream_seeded

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
  manual intervention needed

CONFIGURATION
  access_token (string) required secret: The user's Pocket access token.
  consumer_key (string) required secret: Your application's Consumer Key.
  content_type (string): Select the content type of the items to retrieve.
  detail_type (string): Select the granularity of the information about each item.
  domain (string): Only return items from a particular `domain`.
  favorite (boolean): Retrieve only favorited items.
  search (string): Only return items whose title or url contain the `search` string.
  since (string): Only return items modified since the given timestamp.
  sort (string): Sort retrieved items by the given criteria.
  state (string): Select the state of the items to retrieve.
  tag (string): Return only items tagged with this tag name. Use _untagged_ for retrieving only untagged items.
  secret fields: access_token, consumer_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-pocket

  # Inspect as JSON
  pm connectors inspect source-pocket --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
