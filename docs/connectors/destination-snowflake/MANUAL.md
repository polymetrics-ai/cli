# pm connectors inspect destination-snowflake

```text
NAME
  pm connectors inspect destination-snowflake - Snowflake connector manual

SYNOPSIS
  pm connectors inspect destination-snowflake
  pm connectors inspect destination-snowflake --json
  pm credentials add <name> --connector destination-snowflake [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Snowflake catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/snowflake.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.snowflake.com/en/release-notes

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
  SQL reference: https://docs.snowflake.com/en/sql-reference
  Key pair authentication: https://docs.snowflake.com/en/user-guide/key-pair-auth
  Access control: https://docs.snowflake.com/en/user-guide/security-access-control
  Release notes: https://docs.snowflake.com/en/release-notes
  Snowflake server release notes and feature updates: https://docs.snowflake.com/en/release-notes/new-features
  Snowflake Status: https://status.snowflake.com/

CONFIGURATION
  cdc_deletion_mode (string): Whether to execute CDC deletions as hard deletes (i.e. propagate source deletions to the destination), or soft deletes (i.e. leave a tombstone record in the destination). Defaul...
  credentials (object): Determines the type of authentication that should be used.
  database (string) required: Enter the name of the <a href="https://docs.snowflake.com/en/sql-reference/ddl-database.html#database-schema-share-ddl">database</a> you want to sync data into
  disable_type_dedupe (boolean): Write the legacy "raw tables" format, to enable backwards compatibility with older versions of this connector.
  host (string) required: Enter your Snowflake account's <a href="https://docs.snowflake.com/en/user-guide/admin-account-identifier.html#using-an-account-locator-as-an-identifier">locator</a> (in the for...
  jdbc_url_params (string): Enter the additional properties to pass to the JDBC URL string when connecting to the database (formatted as key=value pairs separated by the symbol &). Example: key1=value1&key...
  raw_data_schema (string): manual intervention needed
  retention_period_days (integer): The number of days of Snowflake Time Travel to enable on the tables. See <a href="https://docs.snowflake.com/en/user-guide/data-time-travel#data-retention-period">Snowflake's do...
  role (string) required: Enter the <a href="https://docs.snowflake.com/en/user-guide/security-access-control-overview.html#roles">role</a> that you want to use to access Snowflake
  schema (string) required: Enter the name of the default <a href="https://docs.snowflake.com/en/sql-reference/ddl-database.html#database-schema-share-ddl">schema</a>
  trim_space (boolean): Whether Snowflake should trim leading and trailing whitespace from fields during data loading. Disable this option if your data contains meaningful leading or trailing whitespac...
  username (string) required: Enter the name of the user you want to use to access the database
  warehouse (string) required: Enter the name of the <a href="https://docs.snowflake.com/en/user-guide/warehouses-overview.html#overview-of-warehouses">warehouse</a> that you want to use as a compute cluster
  secret fields: credentials.password, credentials.private_key, credentials.private_key_password

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-snowflake

  # Inspect as JSON
  pm connectors inspect destination-snowflake --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SQL reference: https://docs.snowflake.com/en/sql-reference
  Key pair authentication: https://docs.snowflake.com/en/user-guide/key-pair-auth
  Access control: https://docs.snowflake.com/en/user-guide/security-access-control
  Release notes: https://docs.snowflake.com/en/release-notes
  Snowflake server release notes and feature updates: https://docs.snowflake.com/en/release-notes/new-features
  Snowflake Status: https://status.snowflake.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
