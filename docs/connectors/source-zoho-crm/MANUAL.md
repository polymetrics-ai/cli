# pm connectors inspect source-zoho-crm

```text
NAME
  pm connectors inspect source-zoho-crm - ZohoCRM connector manual

SYNOPSIS
  pm connectors inspect source-zoho-crm
  pm connectors inspect source-zoho-crm --json
  pm credentials add <name> --connector source-zoho-crm [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  ZohoCRM catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/zohocrm.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.zoho.com/crm/developer/docs/api/v6/

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
  Zoho CRM API: https://www.zoho.com/crm/developer/docs/api/v6/
  Zoho OAuth 2.0: https://www.zoho.com/crm/developer/docs/api/v6/oauth-overview.html
  Zoho CRM API changelog: https://www.zoho.com/crm/developer/docs/api/v6/whats-new.html
  Zoho API limits: https://www.zoho.com/crm/developer/docs/api/v6/api-limits.html

CONFIGURATION
  client_id (string) required secret: OAuth2.0 Client ID
  client_secret (string) required secret: OAuth2.0 Client Secret
  dc_region (string) required: Please choose the region of your Data Center location. More info by this <a href="https://www.zoho.com/crm/developer/docs/api/v2/multi-dc.html">Link</a>
  edition (string) required: Choose your Edition of Zoho CRM to determine API Concurrency Limits
  environment (string) required: Please choose the environment
  refresh_token (string) required secret: OAuth2.0 Refresh Token
  start_datetime ([string null]): ISO 8601, for instance: `YYYY-MM-DD`, `YYYY-MM-DD HH:MM:SS+HH:MM`
  secret fields: client_id, client_secret, refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-zoho-crm

  # Inspect as JSON
  pm connectors inspect source-zoho-crm --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Zoho CRM API: https://www.zoho.com/crm/developer/docs/api/v6/
  Zoho OAuth 2.0: https://www.zoho.com/crm/developer/docs/api/v6/oauth-overview.html
  Zoho CRM API changelog: https://www.zoho.com/crm/developer/docs/api/v6/whats-new.html
  Zoho API limits: https://www.zoho.com/crm/developer/docs/api/v6/api-limits.html

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
