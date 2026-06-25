# pm connectors inspect source-sftp-bulk

```text
NAME
  pm connectors inspect source-sftp-bulk - SFTP Bulk connector manual

SYNOPSIS
  pm connectors inspect source-sftp-bulk
  pm connectors inspect source-sftp-bulk --json
  pm credentials add <name> --connector source-sftp-bulk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SFTP Bulk catalog connector for https://docs.airbyte.com/integrations/sources/sftp-bulk. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: file_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-sftp-bulk:1.9.2 (metadata only; not executed)

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
  SFTP protocol documentation: https://datatracker.ietf.org/doc/html/draft-ietf-secsh-filexfer-02
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/sftp-bulk

CONFIGURATION
  credentials (object) required: Credentials for connecting to the SFTP Server
  delivery_method (object)
  folder_path (string): The directory to search files for sync
  host (string) required: The server host address
  port (integer): The server port
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
  streams (array) required: Each instance of this configuration defines a <a href="https://docs.airbyte.com/cloud/core-concepts#stream">stream</a>. Use this to define which files belong in the stream, thei...
  username (string) required: The server user
  secret fields: credentials.password, credentials.private_key, streams[].format.processing.api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/sftp-bulk

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-sftp-bulk

  # Inspect as JSON
  pm connectors inspect source-sftp-bulk --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SFTP Bulk documentation: https://docs.airbyte.com/integrations/sources/sftp-bulk

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
