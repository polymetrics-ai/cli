# pm connectors inspect source-microsoft-dataverse

```text
NAME
  pm connectors inspect source-microsoft-dataverse - Microsoft Dataverse connector manual

SYNOPSIS
  pm connectors inspect source-microsoft-dataverse
  pm connectors inspect source-microsoft-dataverse --json
  pm credentials add <name> --connector source-microsoft-dataverse [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Microsoft Dataverse catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/microsoftdataverse.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/overview

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
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
  family: custom_go_port
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Microsoft Dataverse Web API: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/overview
  Dataverse authentication: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/authenticate-oauth
  Dataverse API limits: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/api-limits

CONFIGURATION
  client_id (string) required secret: App Registration Client Id
  client_secret_value (string) required secret: App Registration Client Secret
  odata_maxpagesize (integer): Max number of results per page. Default=5000
  tenant_id (string) required secret: Tenant Id of your Microsoft Dataverse Instance
  url (string) required: URL to Microsoft Dataverse API
  secret fields: client_id, client_secret_value, tenant_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-microsoft-dataverse

  # Inspect as JSON
  pm connectors inspect source-microsoft-dataverse --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Microsoft Dataverse Web API: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/overview
  Dataverse authentication: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/authenticate-oauth
  Dataverse API limits: https://learn.microsoft.com/en-us/power-apps/developer/data-platform/api-limits

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
