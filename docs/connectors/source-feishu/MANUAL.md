# pm connectors inspect source-feishu

```text
NAME
  pm connectors inspect source-feishu - Feishu connector manual

SYNOPSIS
  pm connectors inspect source-feishu
  pm connectors inspect source-feishu --json
  pm credentials add <name> --connector source-feishu [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Feishu catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/feishu.svg
  source: official
  review_status: official_verified
  review_url: https://open.feishu.cn/document/server-docs/docs/bitable-v1/bitable-overview

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
  Feishu documentation: https://open.feishu.cn/document/server-docs/docs/bitable-v1/bitable-overview

CONFIGURATION
  app_id (string) required secret: The unique identifier for your application. Found in the Feishu/Lark Developer Console under "Credentials & Basic Info".
  app_secret (string) required secret: The secret key used to verify your application's identity. Found alongside the App ID.
  app_token (string) required secret: The unique identifier of the Bitable (Base). Found in the URL: /base/{app_token}.
  lark_host (string) required: Base URL of the Feishu/Lark Open Platform. Use https://open.feishu.cn for Feishu (China mainland) accounts and https://open.larksuite.com for Lark (international) accounts.
  page_size (number): Number of records per request. Max: 500. Default: 100.
  table_id (string) required: The unique identifier of the table. Found in the URL query parameter table={table_id}.
  secret fields: app_id, app_secret, app_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-feishu

  # Inspect as JSON
  pm connectors inspect source-feishu --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Feishu documentation: https://open.feishu.cn/document/server-docs/docs/bitable-v1/bitable-overview

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
