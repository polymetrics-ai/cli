# pm connectors inspect destination-mysql

```text
NAME
  pm connectors inspect destination-mysql - MySQL connector manual

SYNOPSIS
  pm connectors inspect destination-mysql
  pm connectors inspect destination-mysql --json
  pm credentials add <name> --connector destination-mysql [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  MySQL catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/mysql.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://dev.mysql.com/doc/

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

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
  MySQL documentation: https://dev.mysql.com/doc/
  SQL statement syntax: https://dev.mysql.com/doc/refman/8.0/en/sql-statements.html
  Access control and account management: https://dev.mysql.com/doc/refman/8.0/en/access-control.html
  GRANT statement: https://dev.mysql.com/doc/refman/8.0/en/grant.html
  MySQL Release Notes: https://dev.mysql.com/doc/relnotes/mysql/en/

CONFIGURATION
  database (string) required: Name of the database.
  disable_type_dedupe (boolean): manual intervention needed
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  password (string) secret: Password associated with the username.
  port (integer) required: Port of the database.
  raw_data_schema (string): The database to write raw tables into
  ssl (boolean): Encrypt data using SSL.
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: Username to use to access the database.
  secret fields: password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-mysql

  # Inspect as JSON
  pm connectors inspect destination-mysql --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  MySQL documentation: https://dev.mysql.com/doc/
  SQL statement syntax: https://dev.mysql.com/doc/refman/8.0/en/sql-statements.html
  Access control and account management: https://dev.mysql.com/doc/refman/8.0/en/access-control.html
  GRANT statement: https://dev.mysql.com/doc/refman/8.0/en/grant.html
  MySQL Release Notes: https://dev.mysql.com/doc/relnotes/mysql/en/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
