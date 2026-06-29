# pm connectors inspect destination-postgres

```text
NAME
  pm connectors inspect destination-postgres - Postgres connector manual

SYNOPSIS
  pm connectors inspect destination-postgres
  pm connectors inspect destination-postgres --json
  pm credentials add <name> --connector destination-postgres [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Postgres catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/postgresql.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.postgresql.org/docs/current/

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: generally_available
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
  priority_wave: 1
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  PostgreSQL documentation: https://www.postgresql.org/docs/current/
  SQL commands: https://www.postgresql.org/docs/current/sql-commands.html
  Client authentication: https://www.postgresql.org/docs/current/client-authentication.html
  Database roles and privileges: https://www.postgresql.org/docs/current/user-manag.html
  Release Notes: https://www.postgresql.org/docs/release/

CONFIGURATION
  cdc_deletion_mode (string): Whether to execute CDC deletions as hard deletes (i.e. propagate source deletions to the destination), or soft deletes (i.e. leave a tombstone record in the destination). Defaul...
  database (string) required: Name of the database.
  disable_type_dedupe (boolean): manual intervention needed
  drop_cascade (boolean): Drop tables with CASCADE. WARNING! This will delete all data in all dependent objects (views, etc.). Use with caution. This option is intended for usecases which can easily rebu...
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  password (string) secret: Password associated with the username.
  port (integer) required: Port of the database.
  raw_data_schema (string): manual intervention needed
  schema (string) required: The default schema tables are written. If not specified otherwise, the "public" schema will be used.
  ssl (boolean): Encrypt data using SSL. When activating SSL, please select one of the connection modes.
  ssl_mode (object): manual intervention needed
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  unconstrained_number (boolean): Create numeric columns as unconstrained DECIMAL instead of NUMBER(38, 9). This will allow increased precision in numeric values. (this is disabled by default for backwards compa...
  username (string) required: Username to access the database.
  secret fields: password, ssl_mode.ca_certificate, ssl_mode.client_certificate, ssl_mode.client_key, ssl_mode.client_key_password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-postgres

  # Inspect as JSON
  pm connectors inspect destination-postgres --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  PostgreSQL documentation: https://www.postgresql.org/docs/current/
  SQL commands: https://www.postgresql.org/docs/current/sql-commands.html
  Client authentication: https://www.postgresql.org/docs/current/client-authentication.html
  Database roles and privileges: https://www.postgresql.org/docs/current/user-manag.html
  Release Notes: https://www.postgresql.org/docs/release/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
