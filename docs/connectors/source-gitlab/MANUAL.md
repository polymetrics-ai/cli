# pm connectors inspect source-gitlab

```text
NAME
  pm connectors inspect source-gitlab - Gitlab connector manual

SYNOPSIS
  pm connectors inspect source-gitlab
  pm connectors inspect source-gitlab --json
  pm credentials add <name> --connector source-gitlab [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Gitlab catalog connector for https://docs.airbyte.com/integrations/sources/gitlab. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-gitlab:4.4.31 (metadata only; not executed)

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
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/gitlab

CONFIGURATION
  api_url (string): Please enter your basic URL from GitLab instance.
  credentials (object) required
  groups (string): [DEPRECATED] Space-delimited list of groups. e.g. airbyte.io.
  groups_list (array): List of groups. e.g. airbyte.io.
  num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may hit rate limits. Adjust based on your GitLab instance rate limits.
  projects (string): [DEPRECATED] Space-delimited list of projects. e.g. airbyte.io/documentation meltano/tap-gitlab.
  projects_list (array): Space-delimited list of projects. e.g. airbyte.io/documentation meltano/tap-gitlab.
  start_date (string): The date from which you'd like to replicate data for GitLab API, in the format YYYY-MM-DDT00:00:00Z. Optional. If not set, all data will be replicated. All data generated after ...
  secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/gitlab

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
  Gitlab documentation: https://docs.airbyte.com/integrations/sources/gitlab

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
