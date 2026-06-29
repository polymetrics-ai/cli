# pm connectors inspect source-sharepoint-lists-enterprise

```text
NAME
  pm connectors inspect source-sharepoint-lists-enterprise - SharePoint Lists Enterprise connector manual

SYNOPSIS
  pm connectors inspect source-sharepoint-lists-enterprise
  pm connectors inspect source-sharepoint-lists-enterprise --json
  pm credentials add <name> --connector source-sharepoint-lists-enterprise [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SharePoint Lists Enterprise catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/microsoft-sharepoint.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
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
  family: declarative_http_source
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  SharePoint Lists Enterprise documentation: https://learn.microsoft.com/en-us/graph/api/resources/list?view=graph-rest-1.0

CONFIGURATION
  client_id (string) required secret: Azure AD Application (Client) ID
  client_secret (string) required secret: Azure AD Application Client Secret
  site_id (string) required: SharePoint Site ID. Can be obtained from the site URL or using Microsoft Graph Explorer. Format: {hostname},{site-collection-id},{web-id}
  tenant_id (string) required: Azure AD Tenant (Directory) ID
  secret fields: client_id, client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-sharepoint-lists-enterprise

  # Inspect as JSON
  pm connectors inspect source-sharepoint-lists-enterprise --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SharePoint Lists Enterprise documentation: https://learn.microsoft.com/en-us/graph/api/resources/list?view=graph-rest-1.0

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
