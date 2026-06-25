# pm connectors inspect source-commcare

```text
NAME
  pm connectors inspect source-commcare - Commcare connector manual

SYNOPSIS
  pm connectors inspect source-commcare
  pm connectors inspect source-commcare --json
  pm credentials add <name> --connector source-commcare [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Commcare catalog connector for https://docs.airbyte.com/integrations/sources/commcare. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-commcare:0.1.39 (metadata only; not executed)

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
  CommCare API reference: https://confluence.dimagi.com/display/commcarepublic/CommCare+HQ+APIs
  CommCare authentication: https://confluence.dimagi.com/display/commcarepublic/Authentication
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/commcare

CONFIGURATION
  api_key (string) required secret: Commcare API Key
  app_id (string) required secret: The Application ID we are interested in
  project_space (string): Project Space for commcare
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Only records after this date will be replicated.
  secret fields: api_key, app_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/commcare

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-commcare

  # Inspect as JSON
  pm connectors inspect source-commcare --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Commcare documentation: https://docs.airbyte.com/integrations/sources/commcare

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
