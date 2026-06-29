# pm connectors inspect source-netsuite-enterprise

```text
NAME
  pm connectors inspect source-netsuite-enterprise - Netsuite Enterprise Source connector manual

SYNOPSIS
  pm connectors inspect source-netsuite-enterprise
  pm connectors inspect source-netsuite-enterprise --json
  pm credentials add <name> --connector source-netsuite-enterprise [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Netsuite Enterprise Source catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/netsuite.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/book_N3865324.html

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
  Accessing the Connect Service Using a JDBC Driver: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_3994742720.html
  Connect schema: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_158695828012.html
  Query Language Compliance - SQL Compliance: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_3903316302.html#subsect_163595195498
  NetSuite Release Notes: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/book_N3865324.html

CONFIGURATION
  account_id (string) required: The username which is used to access the database.
  authentication_method (object) required: Configure how to authenticate to Netsuite. Options include username/password or token-based authentication.
  check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
  checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
  concurrency (integer): Maximum number of concurrent queries to the database.
  cursor (object) required: Configures how data is extracted from the database.
  host (string) required: Hostname of the database.
  jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
  port (integer) required: Port of the database.
  role_id (string) required: The username which is used to access the database.
  tunnel_method (object) required: Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
  username (string) required: The username which is used to access the database.
  secret fields: authentication_method.client_secret, authentication_method.oauth2_private_key, authentication_method.password, authentication_method.token_secret, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-netsuite-enterprise

  # Inspect as JSON
  pm connectors inspect source-netsuite-enterprise --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Accessing the Connect Service Using a JDBC Driver: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_3994742720.html
  Connect schema: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_158695828012.html
  Query Language Compliance - SQL Compliance: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_3903316302.html#subsect_163595195498
  NetSuite Release Notes: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/book_N3865324.html

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
