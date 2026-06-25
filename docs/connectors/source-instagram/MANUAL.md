# pm connectors inspect source-instagram

```text
NAME
  pm connectors inspect source-instagram - Instagram connector manual

SYNOPSIS
  pm connectors inspect source-instagram
  pm connectors inspect source-instagram --json
  pm credentials add <name> --connector source-instagram [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Instagram catalog connector for https://docs.airbyte.com/integrations/sources/instagram. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-instagram:4.2.32 (metadata only; not executed)

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
  Instagram Platform Changelog: https://developers.facebook.com/docs/instagram-platform/changelog
  Release notes: https://developers.facebook.com/docs/instagram-api/changelog
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/instagram

CONFIGURATION
  access_token (string) required secret: The value of the access token generated with <b>instagram_basic, instagram_manage_insights, pages_show_list, pages_read_engagement, Instagram Public Content Access</b> permissio...
  client_id (string) secret: The Client ID for your Oauth application
  client_secret (string) secret: The Client Secret for your Oauth application
  num_workers (integer): The number of worker threads to use for the sync.
  start_date (string): The date from which you'd like to replicate data for User Insights, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated. If left blank, the...
  secret fields: access_token, client_id, client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/instagram

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-instagram

  # Inspect as JSON
  pm connectors inspect source-instagram --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Instagram documentation: https://docs.airbyte.com/integrations/sources/instagram

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
