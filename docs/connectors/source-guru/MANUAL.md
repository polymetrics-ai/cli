# pm connectors inspect source-guru

```text
NAME
  pm connectors inspect source-guru - Guru connector manual

SYNOPSIS
  pm connectors inspect source-guru
  pm connectors inspect source-guru --json
  pm credentials add <name> --connector source-guru [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Guru catalog connector for https://docs.airbyte.com/integrations/sources/guru. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-guru:0.0.59 (metadata only; not executed)

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
  Guru API documentation: https://developer.getguru.com/
  Guru authentication: https://developer.getguru.com/docs/authentication
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/guru

CONFIGURATION
  password (string) secret
  search_cards_query (string): Query for searching cards
  start_date (string) required
  team_id (string): Team ID received through response of /teams streams, make sure about access to the team
  username (string) required
  secret fields: password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/guru

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-guru

  # Inspect as JSON
  pm connectors inspect source-guru --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Guru documentation: https://docs.airbyte.com/integrations/sources/guru

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
