# pm connectors inspect source-oracle-enterprise

```text
NAME
  pm connectors inspect source-oracle-enterprise - Oracle connector manual

SYNOPSIS
  pm connectors inspect source-oracle-enterprise
  pm connectors inspect source-oracle-enterprise --json
  pm credentials add <name> --connector source-oracle-enterprise [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Oracle catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/oracle.svg
  source: official
  review_status: official_verified
  review_url: https://docs.oracle.com/en/database/oracle/oracle-database/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: certified

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
  family: database_cdc_source
  priority_wave: 3
  etl_operations: catalog, check, read_cdc, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  cdc_modes: snapshot, oracle_logminer_or_xstream
  cdc_state_fields: scn, snapshot_completed
  conformance: catalog, cdc_checkpoint, cdc_setup_validation, check, delete_semantics, docs_skill, ordering, read_fixture, secret_redaction, snapshot_consistency, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Oracle documentation: https://docs.oracle.com/en/database/oracle/oracle-database/

CONFIGURATION
  additionalProperties (object)
  check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
  checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
  concurrency (integer): Maximum number of concurrent queries to the database.
  connection_data (object) required: The scheme by which to establish a database connection.
  cursor (object) required: Configures how data is extracted from the database.
  encryption (object) required: The encryption method with is used when communicating with the database.
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  max_db_connections (integer): manual intervention needed
  password (string) secret: The password associated with the username.
  port (integer) required: Port of the database. Oracle Corporations recommends the following port numbers: 1521 - Default listening port for client connections to the listener. 2484 - Recommended and off...
  schemas (array): The list of schemas to sync from. Defaults to user. Case sensitive.
  table_filters (array): Inclusion filters for table selection per schema. If no filters are specified for a schema, all tables in that schema will be synced.
  tunnel_method (object) required: Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: The username which is used to access the database.
  secret fields: encryption.ssl_certificate, password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-oracle-enterprise

  # Inspect as JSON
  pm connectors inspect source-oracle-enterprise --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Oracle documentation: https://docs.oracle.com/en/database/oracle/oracle-database/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
