# pm connectors inspect source-jira

```text
NAME
  pm connectors inspect source-jira - Jira connector manual

SYNOPSIS
  pm connectors inspect source-jira
  pm connectors inspect source-jira --json
  pm credentials add <name> --connector source-jira [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Jira catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/jira.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/changelog/#

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
  Changelog: https://developer.atlassian.com/changelog/#
  Jira Platform API Changelog: https://developer.atlassian.com/cloud/jira/platform/changelog/
  Jira Software API Changelog: https://developer.atlassian.com/cloud/jira/software/changelog/
  Jira Cloud Platform API OpenAPI specification: https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json

CONFIGURATION
  credentials (object) required: Choose how to authenticate to Jira.
  domain (string) required: manual intervention needed
  expand_issue_changelog (boolean): (DEPRECATED) Expand the changelog when replicating issues.
  expand_issue_transition (boolean): (DEPRECATED) Expand the transitions when replicating issues.
  issues_stream_expand_with (array): Select fields to Expand the `Issues` stream when replicating with:
  lookback_window_minutes (integer): When set to N, the connector will always refresh resources created within the past N minutes. By default, updated objects that are not newly created are not incrementally synced.
  num_workers (integer): The number of worker threads to use for the sync.
  projects (array): List of Jira project keys to replicate data for, or leave it empty if you want to replicate data for all projects.
  render_fields (boolean): (DEPRECATED) Render issue fields in HTML format in addition to Jira JSON-like format.
  start_date (string): manual intervention needed
  secret fields: credentials.api_token, credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.service_account_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-jira

  # Inspect as JSON
  pm connectors inspect source-jira --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Changelog: https://developer.atlassian.com/changelog/#
  Jira Platform API Changelog: https://developer.atlassian.com/cloud/jira/platform/changelog/
  Jira Software API Changelog: https://developer.atlassian.com/cloud/jira/software/changelog/
  Jira Cloud Platform API OpenAPI specification: https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
