# pm connectors inspect source-linnworks

```text
NAME
  pm connectors inspect source-linnworks - Linnworks connector manual

SYNOPSIS
  pm connectors inspect source-linnworks
  pm connectors inspect source-linnworks --json
  pm credentials add <name> --connector source-linnworks [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Linnworks catalog connector for https://docs.airbyte.com/integrations/sources/linnworks. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-linnworks:0.1.66 (metadata only; not executed)

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
  Linnworks API documentation: https://apps.linnworks.net/Api
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/linnworks

CONFIGURATION
  application_id (string) required: Linnworks Application ID
  application_secret (string) required secret: Linnworks Application Secret
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  token (string) required secret
  secret fields: application_secret, token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/linnworks

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-linnworks

  # Inspect as JSON
  pm connectors inspect source-linnworks --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Linnworks documentation: https://docs.airbyte.com/integrations/sources/linnworks

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
