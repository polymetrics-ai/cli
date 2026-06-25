# pm connectors inspect source-firebolt

```text
NAME
  pm connectors inspect source-firebolt - Firebolt connector manual

SYNOPSIS
  pm connectors inspect source-firebolt
  pm connectors inspect source-firebolt --json
  pm credentials add <name> --connector source-firebolt [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Firebolt catalog connector for https://docs.airbyte.com/integrations/sources/firebolt. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: database_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-firebolt:2.0.39 (metadata only; not executed)

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
  Firebolt API reference: https://docs.firebolt.io/godocs/Overview/api-reference.html
  Firebolt authentication: https://docs.firebolt.io/godocs/Guides/managing-your-organization/service-accounts.html
  Firebolt Status: https://status.firebolt.io/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/firebolt

CONFIGURATION
  account (string) required: Firebolt account to login.
  client_id (string) required: Firebolt service account ID.
  client_secret (string) required secret: Firebolt secret, corresponding to the service account ID.
  database (string) required: The database to connect to.
  engine (string) required: Engine name to connect to.
  host (string): The host name of your Firebolt database.
  secret fields: client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/firebolt

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-firebolt

  # Inspect as JSON
  pm connectors inspect source-firebolt --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Firebolt documentation: https://docs.airbyte.com/integrations/sources/firebolt

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
