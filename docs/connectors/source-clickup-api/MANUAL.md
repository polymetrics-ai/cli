# pm connectors inspect source-clickup-api

```text
NAME
  pm connectors inspect source-clickup-api - ClickUp connector manual

SYNOPSIS
  pm connectors inspect source-clickup-api
  pm connectors inspect source-clickup-api --json
  pm credentials add <name> --connector source-clickup-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  ClickUp catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/clickup.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://clickup.com/api/

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
  ClickUp API reference: https://clickup.com/api/
  ClickUp authentication: https://clickup.com/api/developer-portal/authentication/
  ClickUp rate limits: https://clickup.com/api/developer-portal/rate-limits/
  ClickUp Status: https://status.clickup.com/

CONFIGURATION
  api_token (string) required secret: Every ClickUp API call required authentication. This field is your personal API token. See <a href="https://clickup.com/api/developer-portal/authentication/#personal-token">here...
  include_closed_tasks (boolean): Include or exclude closed tasks. By default, they are excluded. See <a https://clickup.com/api/clickupreference/operation/GetTasks/#!in=query&path=include_closed&t=request">here...
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
  pm connectors inspect source-clickup-api

  # Inspect as JSON
  pm connectors inspect source-clickup-api --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  ClickUp API reference: https://clickup.com/api/
  ClickUp authentication: https://clickup.com/api/developer-portal/authentication/
  ClickUp rate limits: https://clickup.com/api/developer-portal/rate-limits/
  ClickUp Status: https://status.clickup.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
