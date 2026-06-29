# pm connectors inspect destination-mssql-v2

```text
NAME
  pm connectors inspect destination-mssql-v2 - MS SQL Server V2 connector manual

SYNOPSIS
  pm connectors inspect destination-mssql-v2
  pm connectors inspect destination-mssql-v2 --json
  pm credentials add <name> --connector destination-mssql-v2 [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  MS SQL Server V2 catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pm-warehouse.svg
  source: polymetrics
  review_status: polymetrics

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
  MS SQL Server V2 documentation: https://learn.microsoft.com/en-us/sql/sql-server/

CONFIGURATION
  database (string) required: The name of the MSSQL database.
  host (string) required: The host name of the MSSQL database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  load_type (object) required: Specifies the type of load mechanism (e.g., BULK, INSERT) and its associated configuration.
  password (string) secret: The password associated with this username.
  port (integer) required: The port of the MSSQL database.
  schema (string) required: The default schema tables are written to if the source does not specify a namespace. The usual value for this field is "public".
  ssl_method (object) required: The encryption method which is used to communicate with the database.
  user (string) required: The username which is used to access the database.
  secret fields: load_type.shared_access_signature, password, ssl_method.trustStorePassword

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-mssql-v2

  # Inspect as JSON
  pm connectors inspect destination-mssql-v2 --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  MS SQL Server V2 documentation: https://learn.microsoft.com/en-us/sql/sql-server/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
