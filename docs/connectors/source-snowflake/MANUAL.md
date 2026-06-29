# pm connectors inspect source-snowflake

```text
NAME
  pm connectors inspect source-snowflake - Snowflake connector manual

SYNOPSIS
  pm connectors inspect source-snowflake
  pm connectors inspect source-snowflake --json
  pm credentials add <name> --connector source-snowflake [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Snowflake catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/snowflake.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.snowflake.com/

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
  family: database_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

OFFICIAL APPLICATION DOCUMENTATION
  Snowflake documentation: https://docs.snowflake.com/
  Snowflake authentication: https://docs.snowflake.com/en/user-guide/admin-user-management
  Snowflake server release notes and feature updates: https://docs.snowflake.com/en/release-notes/new-features
  Snowflake Status: https://status.snowflake.com/

CONFIGURATION
  check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
  checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
  concurrency (integer): Maximum number of concurrent queries to the database.
  credentials (object)
  cursor (object): Configures how data is extracted from the database.
  database (string) required: manual intervention needed
  host (string) required: The host domain of the snowflake instance (must include the account, region, cloud environment, and end with snowflakecomputing.com).
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  role (string) required: manual intervention needed
  schema (string): The source Snowflake schema tables. Leave empty to access tables from multiple schemas.
  warehouse (string) required: manual intervention needed
  secret fields: credentials.password, credentials.private_key, credentials.private_key_password, credentials.programmatic_access_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-snowflake

  # Inspect as JSON
  pm connectors inspect source-snowflake --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Snowflake documentation: https://docs.snowflake.com/
  Snowflake authentication: https://docs.snowflake.com/en/user-guide/admin-user-management
  Snowflake server release notes and feature updates: https://docs.snowflake.com/en/release-notes/new-features
  Snowflake Status: https://status.snowflake.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
