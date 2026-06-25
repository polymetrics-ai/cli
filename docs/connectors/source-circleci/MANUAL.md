# pm connectors inspect source-circleci

```text
NAME
  pm connectors inspect source-circleci - Circleci connector manual

SYNOPSIS
  pm connectors inspect source-circleci
  pm connectors inspect source-circleci --json
  pm credentials add <name> --connector source-circleci [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Circleci catalog connector for https://docs.airbyte.com/integrations/sources/circleci. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-circleci:0.1.0 (metadata only; not executed)

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
  CircleCI API v2 reference: https://circleci.com/docs/api/v2/
  CircleCI authentication: https://circleci.com/docs/api-developers-guide/#authentication
  CircleCI API rate limits: https://circleci.com/docs/api-developers-guide/#rate-limits
  CircleCI Status: https://status.circleci.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/circleci

CONFIGURATION
  api_key (string) required secret
  job_number (string): Job Number of the workflow for `jobs` stream, Auto fetches from `workflow_jobs` stream, if not configured
  org_id (string) required: The org ID found in `https://app.circleci.com/settings/organization/circleci/xxxxx/overview`
  project_id (string) required: Project ID found in the project settings, Visit `https://app.circleci.com/settings/project/circleci/ORG_SLUG/YYYYY`
  start_date (string) required
  workflow_id (array): Workflow ID of a project pipeline, Could be seen in the URL of pipeline build, Example `https://app.circleci.com/pipelines/circleci/55555xxxxxx/7yyyyyyyyxxxxx/2/workflows/WORKFL...
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/circleci

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-circleci

  # Inspect as JSON
  pm connectors inspect source-circleci --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Circleci documentation: https://docs.airbyte.com/integrations/sources/circleci

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
