# pm connectors inspect source-zendesk-talk

```text
NAME
  pm connectors inspect source-zendesk-talk - Zendesk Talk connector manual

SYNOPSIS
  pm connectors inspect source-zendesk-talk
  pm connectors inspect source-zendesk-talk --json
  pm credentials add <name> --connector source-zendesk-talk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Zendesk Talk catalog connector for https://docs.airbyte.com/integrations/sources/zendesk-talk. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-zendesk-talk:2.0.15 (metadata only; not executed)

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
  Changelog: https://developer.zendesk.com/api-reference/changelog/changelog/
  Developer Updates: https://support.zendesk.com/hc/en-us/sections/4405298889242-Developer-updates
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/zendesk-talk

CONFIGURATION
  credentials (object): Zendesk allows three authentication methods. We recommend using `OAuth2.0` for Airbyte Cloud users and `API token` for Airbyte Open Source users.
  start_date (string) required: The date from which you'd like to replicate data for Zendesk Talk API, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated.
  subdomain (string) required: This is your Zendesk subdomain that can be found in your account URL. For example, in https://{MY_SUBDOMAIN}.zendesk.com/, where MY_SUBDOMAIN is the value of your subdomain.
  subscription_tier (string): Your Zendesk subscription plan tier. This controls API rate limiting behavior — higher tiers have higher rate limits. If unsure, leave as the default (Team) for the most conse...
  secret fields: credentials.access_token, credentials.api_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/zendesk-talk

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-zendesk-talk

  # Inspect as JSON
  pm connectors inspect source-zendesk-talk --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Zendesk Talk documentation: https://docs.airbyte.com/integrations/sources/zendesk-talk

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
