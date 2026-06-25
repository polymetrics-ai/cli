# pm connectors inspect source-mysql

```text
NAME
  pm connectors inspect source-mysql - MySQL connector manual

SYNOPSIS
  pm connectors inspect source-mysql
  pm connectors inspect source-mysql --json
  pm credentials add <name> --connector source-mysql [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  MySQL catalog connector for https://docs.airbyte.com/integrations/sources/mysql. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: database_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-mysql:3.52.3 (metadata only; not executed)

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
  cdc_modes: snapshot, mysql_binlog
  cdc_state_fields: gtid_or_binlog_position, server_id, snapshot_completed
  conformance: catalog, cdc_checkpoint, cdc_setup_validation, check, delete_semantics, docs_skill, ordering, read_fixture, secret_redaction, snapshot_consistency, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  MySQL documentation: https://dev.mysql.com/doc/
  MySQL authentication: https://dev.mysql.com/doc/refman/8.0/en/access-control.html
  MySQL Release Notes: https://dev.mysql.com/doc/relnotes/mysql/en/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/mysql

CONFIGURATION
  check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
  checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
  concurrency (integer): Maximum number of concurrent queries to the database.
  database (string) required: The database name.
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  max_db_connections (integer): Maximum number of concurrent queries to the database. Leave empty to let Airbyte optimize performance.
  password (string) secret: The password associated with the username.
  port (integer) required: Port of the database.
  replication_method (object) required: Configures how data is extracted from the database.
  ssl_mode (object): The encryption method which is used when communicating with the database.
  table_filters (array): Optional filters to include only specific tables from the specified database.
  treat_tinyint1_as_integer (boolean): When enabled, TINYINT(1) columns are emitted as integers instead of booleans.
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: The username which is used to access the database.
  secret fields: password, ssl_mode.ca_certificate, ssl_mode.client_certificate, ssl_mode.client_key, ssl_mode.client_key_password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/mysql

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-mysql

  # Inspect as JSON
  pm connectors inspect source-mysql --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  MySQL documentation: https://docs.airbyte.com/integrations/sources/mysql

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
