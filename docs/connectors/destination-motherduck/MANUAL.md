# pm connectors inspect destination-motherduck

```text
NAME
  pm connectors inspect destination-motherduck - MotherDuck connector manual

SYNOPSIS
  pm connectors inspect destination-motherduck
  pm connectors inspect destination-motherduck --json
  pm credentials add <name> --connector destination-motherduck [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  MotherDuck catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/duckdb.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://motherduck.com/docs

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
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
  family: destination_writer
  priority_wave: 3
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  MotherDuck documentation: https://motherduck.com/docs
  Authentication: https://motherduck.com/docs/key-tasks/authenticating-and-connecting-to-motherduck/
  MotherDuck Version Lifecycle Schedules: https://motherduck.com/docs/troubleshooting/version-lifecycle-schedules/
  MotherDuck Status: https://status.motherduck.com/

CONFIGURATION
  destination_path (string): Path to a .duckdb file or 'md:<DATABASE_NAME>' to connect to a MotherDuck database. If 'md:' is specified without a database name, the default MotherDuck database name ('my_db')...
  motherduck_api_key (string) required secret: API access token to use for authentication to a MotherDuck database.
  schema (string): Database schema name, defaults to 'main' if not specified.
  secret fields: motherduck_api_key

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-motherduck

  # Inspect as JSON
  pm connectors inspect destination-motherduck --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  MotherDuck documentation: https://motherduck.com/docs
  Authentication: https://motherduck.com/docs/key-tasks/authenticating-and-connecting-to-motherduck/
  MotherDuck Version Lifecycle Schedules: https://motherduck.com/docs/troubleshooting/version-lifecycle-schedules/
  MotherDuck Status: https://status.motherduck.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
