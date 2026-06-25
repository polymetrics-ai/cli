# pm connectors inspect source-google-search-console

```text
NAME
  pm connectors inspect source-google-search-console - Google Search Console connector manual

SYNOPSIS
  pm connectors inspect source-google-search-console
  pm connectors inspect source-google-search-console --json
  pm credentials add <name> --connector source-google-search-console [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google Search Console catalog connector for https://docs.airbyte.com/integrations/sources/google-search-console. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-google-search-console:2.1.2 (metadata only; not executed)

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
  Reference: https://developers.google.com/webmaster-tools/v1/api_reference_index
  Google Search Central Blog: https://developers.google.com/search/news
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/google-search-console

CONFIGURATION
  always_use_aggregation_type_auto (boolean): Some search analytics streams fail with a 400 error if the specified `aggregationType` is not supported. This is customer implementation dependent and if this error is encounter...
  authorization (object) required
  custom_reports (string): (DEPRCATED) A JSON array describing the custom reports you want to sync from Google Search Console. See our <a href='https://docs.airbyte.com/integrations/sources/google-search-...
  custom_reports_array (array): You can add your Custom Analytics report by creating one.
  data_state (string): If set to 'final', the returned data will include only finalized, stable data. If set to 'all', fresh data will be included. When using Incremental sync mode, we do not recommen...
  end_date (string): UTC date in the format YYYY-MM-DD. Any data created after this date will not be replicated. Must be greater or equal to the start date field. Leaving this field blank will repli...
  num_workers (integer): The number of worker threads to use for the sync. For more details on Google Search Console rate limits, refer to the <a href="https://developers.google.com/webmaster-tools/limi...
  requests_per_minute (integer): The maximum number of requests per minute for Search Analytics API calls. The default (1200) matches Google's documented maximum quota. If you are experiencing rate limit errors...
  site_urls (array) required: The URLs of the website property attached to your GSC account. Learn more about properties <a href="https://support.google.com/webmasters/answer/34592?hl=en">here</a>.
  start_date (string): UTC date in the format YYYY-MM-DD. Any data before this date will not be replicated.
  secret fields: authorization.access_token, authorization.client_id, authorization.client_secret, authorization.refresh_token, authorization.service_account_info

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/google-search-console

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-google-search-console

  # Inspect as JSON
  pm connectors inspect source-google-search-console --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google Search Console documentation: https://docs.airbyte.com/integrations/sources/google-search-console

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
