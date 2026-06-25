# pm connectors inspect source-merge

```text
NAME
  pm connectors inspect source-merge - Merge connector manual

SYNOPSIS
  pm connectors inspect source-merge
  pm connectors inspect source-merge --json
  pm credentials add <name> --connector source-merge [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Merge catalog connector for https://docs.airbyte.com/integrations/sources/merge. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-merge:0.2.24 (metadata only; not executed)

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
  Merge API reference: https://docs.merge.dev/api-reference/
  Merge authentication: https://docs.merge.dev/basics/authentication/
  Merge rate limits: https://docs.merge.dev/basics/rate-limits/
  Merge Status: https://status.merge.dev/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/merge

CONFIGURATION
  account_token (string) required secret: Link your other integrations with account credentials on accounts section to get account token (ref - https://app.merge.dev/linked-accounts/accounts)
  api_token (string) required secret: API token can be seen at https://app.merge.dev/keys
  start_date (string) required: Date time filter for incremental filter, Specify which date to extract from.
  secret fields: account_token, api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/merge

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-merge

  # Inspect as JSON
  pm connectors inspect source-merge --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Merge documentation: https://docs.airbyte.com/integrations/sources/merge

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
