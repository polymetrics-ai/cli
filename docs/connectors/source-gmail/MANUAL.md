# pm connectors inspect source-gmail

```text
NAME
  pm connectors inspect source-gmail - Gmail connector manual

SYNOPSIS
  pm connectors inspect source-gmail
  pm connectors inspect source-gmail --json
  pm credentials add <name> --connector source-gmail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Gmail catalog connector for https://docs.airbyte.com/integrations/sources/gmail. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-gmail:0.1.5 (metadata only; not executed)

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
  Gmail API reference: https://developers.google.com/gmail/api/reference/rest
  Gmail authentication: https://developers.google.com/gmail/api/auth/about-auth
  Gmail API quotas: https://developers.google.com/gmail/api/reference/quota
  Google Workspace Status: https://www.google.com/appsstatus/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/gmail

CONFIGURATION
  credentials (object) required: Credentials for connecting to the Gmail API
  include_spam_and_trash (boolean): Include drafts/messages from SPAM and TRASH in the results. Defaults to false.
  num_workers (integer): Number of concurrent workers used when syncing. Higher values result in faster syncs but may trigger rate limiting on lower-tier Gmail API quotas. The default works well for mos...
  start_date (string): UTC date and time in the format YYYY-MM-DDTHH:MM:SSZ. Only messages, threads, and drafts received on or after this date will be replicated. If not set, all historical data will ...
  secret fields: credentials.client_id, credentials.client_refresh_token, credentials.client_secret, credentials.service_account_info

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/gmail

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-gmail

  # Inspect as JSON
  pm connectors inspect source-gmail --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Gmail documentation: https://docs.airbyte.com/integrations/sources/gmail

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
