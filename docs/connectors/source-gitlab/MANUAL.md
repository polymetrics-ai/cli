# pm connectors inspect source-gitlab

```text
NAME
  pm connectors inspect source-gitlab - Gitlab connector manual

SYNOPSIS
  pm connectors inspect source-gitlab
  pm connectors inspect source-gitlab --json
  pm credentials add <name> --connector source-gitlab [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Gitlab catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/gitlab.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.gitlab.com/ee/api/rest/deprecations.html

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  API reference: https://docs.gitlab.com/ee/api/rest/
  Future REST API deprecations and removals: https://docs.gitlab.com/ee/api/rest/deprecations.html
  GitLab API OpenAPI specification: https://docs.gitlab.com/ee/api/openapi/openapi.yaml

CONFIGURATION
  api_url (string): Please enter your basic URL from GitLab instance.
  credentials (object) required
  groups (string): manual intervention needed
  groups_list (array): manual intervention needed
  num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may hit rate limits. Adjust based on your GitLab instance rate limits.
  projects (string): manual intervention needed
  projects_list (array): manual intervention needed
  start_date (string): The date from which you'd like to replicate data for GitLab API, in the format YYYY-MM-DDT00:00:00Z. Optional. If not set, all data will be replicated. All data generated after ...
  secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-gitlab

  # Inspect as JSON
  pm connectors inspect source-gitlab --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  API reference: https://docs.gitlab.com/ee/api/rest/
  Future REST API deprecations and removals: https://docs.gitlab.com/ee/api/rest/deprecations.html
  GitLab API OpenAPI specification: https://docs.gitlab.com/ee/api/openapi/openapi.yaml

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
