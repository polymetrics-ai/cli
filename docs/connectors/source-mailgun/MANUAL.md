# pm connectors inspect source-mailgun

```text
NAME
  pm connectors inspect source-mailgun - Mailgun connector manual

SYNOPSIS
  pm connectors inspect source-mailgun
  pm connectors inspect source-mailgun --json
  pm credentials add <name> --connector source-mailgun [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Mailgun catalog connector for https://docs.airbyte.com/integrations/sources/mailgun. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-mailgun:0.3.54 (metadata only; not executed)

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
  Mailgun API reference: https://documentation.mailgun.com/en/latest/api_reference.html
  Mailgun authentication: https://documentation.mailgun.com/en/latest/api-intro.html#authentication
  Mailgun rate limits: https://documentation.mailgun.com/en/latest/api-intro.html#rate-limiting
  Mailgun Status: https://status.mailgun.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/mailgun

CONFIGURATION
  domain_region (string): Domain region code. 'EU' or 'US' are possible values. The default is 'US'.
  private_key (string) required secret: Primary account API key to access your Mailgun data.
  start_date (string): UTC date and time in the format 2020-10-01 00:00:00. Any data before this date will not be replicated. If omitted, defaults to 3 days ago.
  secret fields: private_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/mailgun

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-mailgun

  # Inspect as JSON
  pm connectors inspect source-mailgun --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Mailgun documentation: https://docs.airbyte.com/integrations/sources/mailgun

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
