# pm connectors inspect source-sftp

```text
NAME
  pm connectors inspect source-sftp - SFTP connector manual

SYNOPSIS
  pm connectors inspect source-sftp
  pm connectors inspect source-sftp --json
  pm credentials add <name> --connector source-sftp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SFTP catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/sftp.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: file_go
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
  family: file_object_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  manual intervention needed

CONFIGURATION
  credentials (object): The server authentication method
  file_pattern (string): The regular expression to specify files for sync in a chosen Folder Path
  file_types (string): Coma separated file types. Currently only 'csv' and 'json' types are supported.
  folder_path (string): The directory to search files for sync
  host (string) required: The server host address
  port (integer) required: The server port
  user (string) required: The server user
  secret fields: credentials.auth_ssh_key, credentials.auth_user_password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-sftp

  # Inspect as JSON
  pm connectors inspect source-sftp --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
