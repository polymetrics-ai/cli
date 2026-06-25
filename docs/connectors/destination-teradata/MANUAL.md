# pm connectors inspect destination-teradata

```text
NAME
  pm connectors inspect destination-teradata - Teradata Vantage connector manual

SYNOPSIS
  pm connectors inspect destination-teradata
  pm connectors inspect destination-teradata --json
  pm credentials add <name> --connector destination-teradata [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Teradata Vantage catalog connector for https://docs.airbyte.com/integrations/destinations/teradata. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: destination
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: destination_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/destination-teradata:1.0.0 (metadata only; not executed)

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
  Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/teradata

CONFIGURATION
  disable_type_dedupe (boolean): Disable Writing Final Tables. WARNING! The data format in _airbyte_data is likely stable but there are no guarantees that other metadata columns will remain the same in future v...
  drop_cascade (boolean): Drop tables with CASCADE. WARNING! This will delete all data in all dependent objects (views, etc.). Use with caution. This option is intended for usecases which can easily rebu...
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  logmech (object)
  query_band (string): Defines the custom session query band using name-value pairs. For example, 'org=Finance;report=Fin123;'
  raw_data_schema (string): The database to write raw tables into
  schema (string): The default schema tables are written to if the source does not specify a namespace. The usual value for this field is "public".
  ssl (boolean): Encrypt data using SSL. When activating SSL, please select one of the SSL modes.
  ssl_mode (object): SSL connection modes. <b>disable</b> - Chose this mode to disable encryption of communication between Airbyte and destination database <b>allow</b> - Chose this mode to enable e...
  secret fields: logmech.password, ssl_mode.ssl_ca_certificate

SYNC MODES
  supported sync modes: append, append_dedup, overwrite
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/destinations/teradata

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect destination-teradata

  # Inspect as JSON
  pm connectors inspect destination-teradata --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Teradata Vantage documentation: https://docs.airbyte.com/integrations/destinations/teradata

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
