# pm connectors inspect source-trello

```text
NAME
  pm connectors inspect source-trello - Trello connector manual

SYNOPSIS
  pm connectors inspect source-trello
  pm connectors inspect source-trello --json
  pm credentials add <name> --connector source-trello [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Trello catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/trello.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/cloud/trello/rest/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Trello REST API: https://developer.atlassian.com/cloud/trello/rest/
  Trello authentication: https://developer.atlassian.com/cloud/trello/guides/rest-api/authorization/
  Trello rate limits: https://developer.atlassian.com/cloud/trello/guides/rest-api/rate-limits/

CONFIGURATION
  board_ids (array): IDs of the boards to replicate data from. If left empty, data from all boards to which you have access will be replicated. Please note that this is not the 8-character ID in the...
  key (string) required secret: Trello API key. See the <a href="https://developer.atlassian.com/cloud/trello/guides/rest-api/authorization/#using-basic-oauth">docs</a> for instructions on how to generate it.
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  token (string) required secret: Trello API token. See the <a href="https://developer.atlassian.com/cloud/trello/guides/rest-api/authorization/#using-basic-oauth">docs</a> for instructions on how to generate it.
  secret fields: key, token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-trello

  # Inspect as JSON
  pm connectors inspect source-trello --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Trello REST API: https://developer.atlassian.com/cloud/trello/rest/
  Trello authentication: https://developer.atlassian.com/cloud/trello/guides/rest-api/authorization/
  Trello rate limits: https://developer.atlassian.com/cloud/trello/guides/rest-api/rate-limits/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
