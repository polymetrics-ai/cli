# pm connectors inspect destination-oracle

```text
NAME
  pm connectors inspect destination-oracle - Oracle connector manual

SYNOPSIS
  pm connectors inspect destination-oracle
  pm connectors inspect destination-oracle --json
  pm credentials add <name> --connector destination-oracle [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Oracle catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/oracle.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.oracle.com/en/database/

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
  Oracle Database documentation: https://docs.oracle.com/en/database/
  SQL language reference: https://docs.oracle.com/en/database/oracle/oracle-database/19/sqlrf/
  Database authentication: https://docs.oracle.com/en/database/oracle/oracle-database/19/dbseg/configuring-authentication.html
  Managing security: https://docs.oracle.com/en/database/oracle/oracle-database/19/dbseg/
  Oracle Database Release Notes: https://docs.oracle.com/en/database/oracle/oracle-database/

CONFIGURATION
  encryption (object): The encryption method which is used when communicating with the database.
  host (string) required: The hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  password (string) secret: The password associated with the username.
  port (integer) required: The port of the database.
  raw_data_schema (string): manual intervention needed
  schema (string): manual intervention needed
  sid (string) required: The System Identifier uniquely distinguishes the instance from any other instance on the same computer.
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: The username to access the database. This user must have CREATE USER privileges in the database.
  secret fields: encryption.ssl_certificate, password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-oracle

  # Inspect as JSON
  pm connectors inspect destination-oracle --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Oracle Database documentation: https://docs.oracle.com/en/database/
  SQL language reference: https://docs.oracle.com/en/database/oracle/oracle-database/19/sqlrf/
  Database authentication: https://docs.oracle.com/en/database/oracle/oracle-database/19/dbseg/configuring-authentication.html
  Managing security: https://docs.oracle.com/en/database/oracle/oracle-database/19/dbseg/
  Oracle Database Release Notes: https://docs.oracle.com/en/database/oracle/oracle-database/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
