# pm connectors inspect source-mailersend

```text
NAME
  pm connectors inspect source-mailersend - MailerSend connector manual

SYNOPSIS
  pm connectors inspect source-mailersend
  pm connectors inspect source-mailersend --json
  pm credentials add <name> --connector source-mailersend [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  MailerSend catalog connector for https://docs.airbyte.com/integrations/sources/mailersend. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-mailersend:0.2.25 (metadata only; not executed)

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
  MailerSend API reference: https://developers.mailersend.com/
  MailerSend authentication: https://developers.mailersend.com/#authentication
  MailerSend rate limits: https://developers.mailersend.com/#rate-limits
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/mailersend

CONFIGURATION
  api_token (string) required secret: Your API Token. See <a href="https://www.mailersend.com/help/managing-api-tokens">here</a>.
  domain_id (string) required: The domain entity in mailersend
  start_date (number): Timestamp is assumed to be UTC.
  secret fields: api_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/mailersend

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-mailersend

  # Inspect as JSON
  pm connectors inspect source-mailersend --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  MailerSend documentation: https://docs.airbyte.com/integrations/sources/mailersend

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
