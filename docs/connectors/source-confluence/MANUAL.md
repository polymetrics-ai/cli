# pm connectors inspect source-confluence

```text
NAME
  pm connectors inspect source-confluence - Confluence connector manual

SYNOPSIS
  pm connectors inspect source-confluence
  pm connectors inspect source-confluence --json
  pm credentials add <name> --connector source-confluence [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Confluence catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/confluence.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Confluence Cloud REST API: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/
  Confluence authentication: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/#authentication
  Confluence rate limits: https://developer.atlassian.com/cloud/confluence/rate-limiting/
  Atlassian Status: https://status.atlassian.com/

CONFIGURATION
  api_token (string) required secret: Please follow the Jira confluence for generating an API token: <a href="https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/">gener...
  domain_name (string) required: Your Confluence domain name
  email (string) required: Your Confluence login email
  secret fields: api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-confluence

  # Inspect as JSON
  pm connectors inspect source-confluence --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Confluence Cloud REST API: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/
  Confluence authentication: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/#authentication
  Confluence rate limits: https://developer.atlassian.com/cloud/confluence/rate-limiting/
  Atlassian Status: https://status.atlassian.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
