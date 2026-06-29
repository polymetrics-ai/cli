# pm connectors inspect destination-databricks

```text
NAME
  pm connectors inspect destination-databricks - Databricks Lakehouse connector manual

SYNOPSIS
  pm connectors inspect destination-databricks
  pm connectors inspect destination-databricks --json
  pm credentials add <name> --connector destination-databricks [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Databricks Lakehouse catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/databricks.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.databricks.com/release-notes/index.html

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
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
  priority_wave: 3
  etl_operations: catalog, check, write_append, write_dedup, write_overwrite
  reverse_etl_operations: none until native write conformance passes
  conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

OFFICIAL APPLICATION DOCUMENTATION
  SQL reference: https://docs.databricks.com/sql/language-manual/index.html
  Authentication: https://docs.databricks.com/dev-tools/auth.html
  Access control: https://docs.databricks.com/security/access-control/index.html
  Release notes: https://docs.databricks.com/release-notes/index.html
  Databricks Status: https://status.databricks.com/

CONFIGURATION
  accept_terms (boolean) required: You must agree to the Databricks JDBC Driver <a href="https://databricks.com/jdbc-odbc-driver-license">Terms & Conditions</a> to use this connector.
  authentication (object) required: Authentication mechanism for Staging files and running queries
  database (string) required: The name of the unity catalog for the database
  hostname (string) required: Databricks Cluster Server Hostname.
  http_path (string) required: Databricks Cluster HTTP Path.
  port (string): Databricks Cluster Port.
  purge_staging_data (boolean): Default to 'true'. Switch it to 'false' for debugging purpose.
  raw_schema_override (string): manual intervention needed
  schema (string): The default schema tables are written. If not specified otherwise, the "default" will be used.
  secret fields: authentication.personal_access_token, authentication.secret

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-databricks

  # Inspect as JSON
  pm connectors inspect destination-databricks --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SQL reference: https://docs.databricks.com/sql/language-manual/index.html
  Authentication: https://docs.databricks.com/dev-tools/auth.html
  Access control: https://docs.databricks.com/security/access-control/index.html
  Release notes: https://docs.databricks.com/release-notes/index.html
  Databricks Status: https://status.databricks.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
