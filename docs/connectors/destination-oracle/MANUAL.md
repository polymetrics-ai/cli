# pm connectors inspect destination-oracle

```text
NAME
  pm connectors inspect destination-oracle - Oracle connector manual

SYNOPSIS
  pm connectors inspect destination-oracle
  pm connectors inspect destination-oracle --json
  pm credentials add <name> --connector destination-oracle [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Oracle catalog connector for https://docs.airbyte.com/integrations/destinations/oracle. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-oracle:1.0.0 (metadata only; not executed)

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
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/oracle

CONFIGURATION
  encryption (object): The encryption method which is used when communicating with the database.
  host (string) required: The hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  password (string) secret: The password associated with the username.
  port (integer) required: The port of the database.
  raw_data_schema (string): The schema to write raw tables into (default: airbyte_internal)
  schema (string): The default schema is used as the target schema for all statements issued from the connection that do not explicitly specify a schema name. The usual value for this field is "ai...
  sid (string) required: The System Identifier uniquely distinguishes the instance from any other instance on the same computer.
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: The username to access the database. This user must have CREATE USER privileges in the database.
  secret fields: encryption.ssl_certificate, password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/oracle

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
  Oracle documentation: https://docs.airbyte.com/integrations/destinations/oracle

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
