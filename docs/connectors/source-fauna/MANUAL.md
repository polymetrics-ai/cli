# pm connectors inspect source-fauna

```text
NAME
  pm connectors inspect source-fauna - Fauna connector manual

SYNOPSIS
  pm connectors inspect source-fauna
  pm connectors inspect source-fauna --json
  pm credentials add <name> --connector source-fauna [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Fauna catalog connector for https://docs.airbyte.com/integrations/sources/fauna. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: database_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-fauna:0.1.9 (metadata only; not executed)

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
  family: database_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

OFFICIAL APPLICATION DOCUMENTATION
  Fauna API reference: https://docs.fauna.com/fauna/current/api/
  Fauna authentication: https://docs.fauna.com/fauna/current/security/
  Fauna Status: https://status.fauna.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/fauna

CONFIGURATION
  collection (object): Settings for the Fauna Collection.
  domain (string) required: Domain of Fauna to query. Defaults db.fauna.com. See <a href=https://docs.fauna.com/fauna/current/learn/understanding/region_groups#how-to-use-region-groups>the docs</a>.
  port (integer) required: Endpoint port.
  scheme (string) required: URL scheme.
  secret (string) required secret: Fauna secret, used when authenticating with the database.
  secret fields: secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/fauna

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-fauna

  # Inspect as JSON
  pm connectors inspect source-fauna --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Fauna documentation: https://docs.airbyte.com/integrations/sources/fauna

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
