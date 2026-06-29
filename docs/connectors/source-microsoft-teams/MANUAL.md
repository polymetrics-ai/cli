# pm connectors inspect source-microsoft-teams

```text
NAME
  pm connectors inspect source-microsoft-teams - Microsoft Teams connector manual

SYNOPSIS
  pm connectors inspect source-microsoft-teams
  pm connectors inspect source-microsoft-teams --json
  pm credentials add <name> --connector source-microsoft-teams [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Microsoft Teams catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/microsoft-teams.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

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
  Microsoft Teams API: https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview
  Microsoft Graph authentication: https://learn.microsoft.com/en-us/graph/auth/
  Microsoft Graph throttling: https://learn.microsoft.com/en-us/graph/throttling
  Microsoft 365 Status: https://status.office365.com/

CONFIGURATION
  credentials (object): Choose how to authenticate to Microsoft
  period (string) required: Specifies the length of time over which the Team Device Report stream is aggregated. The supported values are: D7, D30, D90, and D180.
  secret fields: credentials.client_secret, credentials.refresh_token, credentials.tenant_id

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-microsoft-teams

  # Inspect as JSON
  pm connectors inspect source-microsoft-teams --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Microsoft Teams API: https://learn.microsoft.com/en-us/graph/api/resources/teams-api-overview
  Microsoft Graph authentication: https://learn.microsoft.com/en-us/graph/auth/
  Microsoft Graph throttling: https://learn.microsoft.com/en-us/graph/throttling
  Microsoft 365 Status: https://status.office365.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
