# pm connectors inspect source-box-data-extract

```text
NAME
  pm connectors inspect source-box-data-extract - Box Data Extract connector manual

SYNOPSIS
  pm connectors inspect source-box-data-extract
  pm connectors inspect source-box-data-extract --json
  pm credentials add <name> --connector source-box-data-extract [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Box Data Extract catalog connector for https://docs.airbyte.com/integrations/sources/box-data-extract. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-box-data-extract:0.1.13 (metadata only; not executed)

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
  Box Platform API reference: https://developer.box.com/reference/
  Box authentication guide: https://developer.box.com/guides/authentication/
  Box API rate limits: https://developer.box.com/guides/api-calls/permissions-and-errors/rate-limits/
  Box Platform Status: https://status.box.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/box-data-extract

CONFIGURATION
  ask_ai_prompt (string): Prompt to use in Ask AI Stream.
  box_folder_id (string) required: Folder to retreive data from.
  box_subject_id (string) required: If subject type is "enterprise", use your enterprise ID If subject type is "user", use the user id to login as.
  box_subject_type (string) required: Represents the type of user to login as ("user" or "enterprise"). Enterprise will login with the application service account. User will login with the user if app can impersonat...
  client_id (string) required: You Box App client ID. Find yours in the <a href="https://app.box.com/developers/console">developer console</a>.
  client_secret (string) required secret: You Box App client secret. Find yours in the <a href="https://app.box.com/developers/console">developer console</a>.
  extract_ai_prompt (string): Prompt to use in Extract AI Stream.
  extract_structured_ai_fields (string): Prompt to use in Extract Strctured AI Stream.
  is_recursive (boolean) required: Read the folders recursively.
  secret fields: client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/box-data-extract

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-box-data-extract

  # Inspect as JSON
  pm connectors inspect source-box-data-extract --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Box Data Extract documentation: https://docs.airbyte.com/integrations/sources/box-data-extract

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
