# pm connectors inspect destination-redshift

```text
NAME
  pm connectors inspect destination-redshift - Redshift connector manual

SYNOPSIS
  pm connectors inspect destination-redshift
  pm connectors inspect destination-redshift --json
  pm credentials add <name> --connector destination-redshift [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Redshift catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/redshift.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.aws.amazon.com/redshift/latest/mgmt/cluster-versions.html

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
  SQL reference: https://docs.aws.amazon.com/redshift/latest/dg/c_SQL_reference.html
  Database authentication: https://docs.aws.amazon.com/redshift/latest/mgmt/generating-user-credentials.html
  Access control: https://docs.aws.amazon.com/redshift/latest/dg/r_GRANT.html
  Cluster versions: https://docs.aws.amazon.com/redshift/latest/mgmt/cluster-versions.html
  Release notes: https://docs.aws.amazon.com/redshift/latest/mgmt/release-notes.html
  Quotas and limits: https://docs.aws.amazon.com/redshift/latest/mgmt/amazon-redshift-limits.html
  AWS Service Health Dashboard: https://health.aws.amazon.com/health/status

CONFIGURATION
  database (string) required: Enter the name of the <a href="https://docs.aws.amazon.com/redshift/latest/dg/r_CREATE_DATABASE.html">database</a> you want to sync data into
  drop_cascade (boolean): WARNING! This will delete all data in all dependent objects (views, etc.) including during schema evolution of columns. Use with caution. This option is intended for usecases wh...
  host (string) required: Enter your Redshift Cluster Endpoint (must include the cluster-id, region and end with .redshift.amazonaws.com)
  jdbc_url_params (string): Enter the additional properties to pass to the JDBC URL string when connecting to the database (formatted as key=value pairs separated by the symbol &). Example: key1=value1&key...
  password (string) required secret: Enter the password associated with the username.
  port (integer) required: Enter your Database Port
  schema (string) required: Enter the name of the default <a href="https://docs.aws.amazon.com/redshift/latest/dg/r_Schemas_and_tables.html">schema</a> tables are written to if the source does not specify ...
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  uploading_method (object): The way data will be uploaded to Redshift.
  username (string) required: Enter the name of the user you want to use to access the database
  secret fields: password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password, uploading_method.access_key_id, uploading_method.secret_access_key

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-redshift

  # Inspect as JSON
  pm connectors inspect destination-redshift --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SQL reference: https://docs.aws.amazon.com/redshift/latest/dg/c_SQL_reference.html
  Database authentication: https://docs.aws.amazon.com/redshift/latest/mgmt/generating-user-credentials.html
  Access control: https://docs.aws.amazon.com/redshift/latest/dg/r_GRANT.html
  Cluster versions: https://docs.aws.amazon.com/redshift/latest/mgmt/cluster-versions.html
  Release notes: https://docs.aws.amazon.com/redshift/latest/mgmt/release-notes.html
  Quotas and limits: https://docs.aws.amazon.com/redshift/latest/mgmt/amazon-redshift-limits.html
  AWS Service Health Dashboard: https://health.aws.amazon.com/health/status

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
