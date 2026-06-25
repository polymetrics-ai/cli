# pm connectors inspect source-google-directory

```text
NAME
  pm connectors inspect source-google-directory - Google Directory connector manual

SYNOPSIS
  pm connectors inspect source-google-directory
  pm connectors inspect source-google-directory --json
  pm credentials add <name> --connector source-google-directory [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google Directory catalog connector for https://docs.airbyte.com/integrations/sources/google-directory. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-google-directory:0.2.45 (metadata only; not executed)

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
  family: custom_go_port
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Google Directory API reference: https://developers.google.com/admin-sdk/directory/reference/rest
  Google Directory authentication: https://developers.google.com/admin-sdk/directory/v1/guides/authorizing
  Google Workspace Status: https://www.google.com/appsstatus/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/google-directory

CONFIGURATION
  credentials (object): Google APIs use the OAuth 2.0 protocol for authentication and authorization. The Source supports <a href="https://developers.google.com/identity/protocols/oauth2#webserver" targ...
  secret fields: credentials.client_id, credentials.client_secret, credentials.credentials_json, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/google-directory

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-google-directory

  # Inspect as JSON
  pm connectors inspect source-google-directory --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google Directory documentation: https://docs.airbyte.com/integrations/sources/google-directory

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
