# pm connectors inspect source-oracle

```text
NAME
  pm connectors inspect source-oracle - Oracle DB connector manual

SYNOPSIS
  pm connectors inspect source-oracle
  pm connectors inspect source-oracle --json
  pm credentials add <name> --connector source-oracle [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Oracle DB catalog connector for https://docs.airbyte.com/integrations/sources/oracle. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: database_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-oracle:0.5.8 (metadata only; not executed)

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
  family: database_cdc_source
  priority_wave: 3
  etl_operations: catalog, check, read_cdc, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  cdc_modes: snapshot, oracle_logminer_or_xstream
  cdc_state_fields: scn, snapshot_completed
  conformance: catalog, cdc_checkpoint, cdc_setup_validation, check, delete_semantics, docs_skill, ordering, read_fixture, secret_redaction, snapshot_consistency, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Oracle Database documentation: https://docs.oracle.com/en/database/
  Oracle authentication: https://docs.oracle.com/en/database/oracle/oracle-database/19/dbseg/introduction-to-oracle-database-security.html
  Oracle Database Release Notes: https://docs.oracle.com/en/database/oracle/oracle-database/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/oracle

CONFIGURATION
  connection_data (object): Connect data that will be used for DB connection
  encryption (object): The encryption method with is used when communicating with the database.
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  password (string) secret: The password associated with the username.
  port (integer) required: Port of the database. Oracle Corporations recommends the following port numbers: 1521 - Default listening port for client connections to the listener. 2484 - Recommended and off...
  schemas (array): The list of schemas to sync from. Defaults to user. Case sensitive.
  tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: The username which is used to access the database.
  secret fields: encryption.ssl_certificate, password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/oracle

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-oracle

  # Inspect as JSON
  pm connectors inspect source-oracle --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Oracle DB documentation: https://docs.airbyte.com/integrations/sources/oracle

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
