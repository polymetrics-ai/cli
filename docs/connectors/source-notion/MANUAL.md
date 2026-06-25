# pm connectors inspect source-notion

```text
NAME
  pm connectors inspect source-notion - Notion connector manual

SYNOPSIS
  pm connectors inspect source-notion
  pm connectors inspect source-notion --json
  pm credentials add <name> --connector source-notion [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Notion catalog connector for https://docs.airbyte.com/integrations/sources/notion. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-notion:4.0.13 (metadata only; not executed)

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
  Changes by version: https://developers.notion.com/reference/changes-by-version
  Changelog: https://developers.notion.com/page/changelog
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/notion

CONFIGURATION
  credentials (object): Choose either OAuth (recommended for Airbyte Cloud) or Access Token. See our <a href='https://docs.airbyte.com/integrations/sources/notion#setup-guide'>docs</a> for more informa...
  num_workers (integer): Number of worker threads to use for the sync. Higher values can speed up large syncs but may increase rate-limit pressure against Notion's limit of approximately three requests ...
  start_date (string): UTC date and time in the format YYYY-MM-DDTHH:MM:SS.000Z. During incremental sync, any data generated before this date will not be replicated. If left blank, the start date will...
  secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/notion

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-notion

  # Inspect as JSON
  pm connectors inspect source-notion --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Notion documentation: https://docs.airbyte.com/integrations/sources/notion

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
