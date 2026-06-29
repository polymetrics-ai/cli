# pm connectors inspect source-teradata

```text
NAME
  pm connectors inspect source-teradata - Teradata connector manual

SYNOPSIS
  pm connectors inspect source-teradata
  pm connectors inspect source-teradata --json
  pm credentials add <name> --connector source-teradata [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Teradata catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/teradata.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.teradata.com/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: database_go
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
  family: database_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

OFFICIAL APPLICATION DOCUMENTATION
  Teradata documentation: https://docs.teradata.com/

CONFIGURATION
  database (string) required: Name of the database.
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  password (string) secret: Password associated with the username.
  port (integer): Port of the database.
  replication_method (string): Replication method to use for extracting data from the database. STANDARD replication requires no setup on the DB side but will not be able to represent deletions incrementally....
  ssl (boolean): Encrypt data using SSL. When activating SSL, please select one of the connection modes.
  ssl_mode (object): manual intervention needed
  username (string) required: Username to use to access the database.
  secret fields: password, ssl_mode.ssl_ca_certificate

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-teradata

  # Inspect as JSON
  pm connectors inspect source-teradata --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Teradata documentation: https://docs.teradata.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
