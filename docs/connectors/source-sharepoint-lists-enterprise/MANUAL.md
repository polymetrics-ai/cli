# pm connectors inspect source-sharepoint-lists-enterprise

```text
NAME
  pm connectors inspect source-sharepoint-lists-enterprise - SharePoint Lists Enterprise connector manual

SYNOPSIS
  pm connectors inspect source-sharepoint-lists-enterprise
  pm connectors inspect source-sharepoint-lists-enterprise --json
  pm credentials add <name> --connector source-sharepoint-lists-enterprise [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SharePoint Lists Enterprise catalog connector for https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-lists-enterprise. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-sharepoint-lists-enterprise:0.1.0 (metadata only; not executed)

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
  No upstream application documentation URL was listed in the imported connector registry.
  Airbyte connector documentation: https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-lists-enterprise

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
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-lists-enterprise

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
  SharePoint Lists Enterprise documentation: https://docs.airbyte.com/integrations/enterprise-connectors/source-sharepoint-lists-enterprise

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
