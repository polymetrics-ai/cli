# pm connectors inspect destination-yellowbrick

```text
NAME
  pm connectors inspect destination-yellowbrick - Yellowbrick connector manual

SYNOPSIS
  pm connectors inspect destination-yellowbrick
  pm connectors inspect destination-yellowbrick --json
  pm credentials add <name> --connector destination-yellowbrick [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Yellowbrick catalog connector for https://docs.airbyte.com/integrations/destinations/yellowbrick. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-yellowbrick:0.0.4 (metadata only; not executed)

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
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/yellowbrick

CONFIGURATION
  database (string) required: Name of the database.
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  password (string) secret: Password associated with the username.
  port (integer) required: Port of the database.
  schema (string) required: The default schema tables are written to if the source does not specify a namespace. The usual value for this field is "public".
  ssl (boolean): Encrypt data using SSL. When activating SSL, please select one of the connection modes.
  ssl_mode (object): SSL connection modes. <b>disable</b> - Chose this mode to disable encryption of communication between Airbyte and destination database <b>allow</b> - Chose this mode to enable e...
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: Username to use to access the database.
  secret fields: password, ssl_mode.ca_certificate, ssl_mode.client_certificate, ssl_mode.client_key, ssl_mode.client_key_password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/yellowbrick

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-yellowbrick

  # Inspect as JSON
  pm connectors inspect destination-yellowbrick --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Yellowbrick documentation: https://docs.airbyte.com/integrations/destinations/yellowbrick

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
