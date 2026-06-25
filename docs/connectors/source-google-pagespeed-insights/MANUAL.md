# pm connectors inspect source-google-pagespeed-insights

```text
NAME
  pm connectors inspect source-google-pagespeed-insights - Google PageSpeed Insights connector manual

SYNOPSIS
  pm connectors inspect source-google-pagespeed-insights
  pm connectors inspect source-google-pagespeed-insights --json
  pm credentials add <name> --connector source-google-pagespeed-insights [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Google PageSpeed Insights catalog connector for https://docs.airbyte.com/integrations/sources/google-pagespeed-insights. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-google-pagespeed-insights:0.2.52 (metadata only; not executed)

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
  PageSpeed Insights API: https://developers.google.com/speed/docs/insights/v5/get-started
  PageSpeed Insights quotas: https://developers.google.com/speed/docs/insights/v5/get-started#quota
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/google-pagespeed-insights

CONFIGURATION
  api_key (string) secret: Google PageSpeed API Key. See <a href="https://developers.google.com/speed/docs/insights/v5/get-started#APIKey">here</a>. The key is optional - however the API is heavily rate l...
  categories (array) required: Defines which Lighthouse category to run. One or many of: "accessibility", "best-practices", "performance", "pwa", "seo".
  strategies (array) required: The analyses strategy to use. Either "desktop" or "mobile".
  urls (array) required: The URLs to retrieve pagespeed information from. The connector will attempt to sync PageSpeed reports for all the defined URLs. Format: https://(www.)url.domain
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/google-pagespeed-insights

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-google-pagespeed-insights

  # Inspect as JSON
  pm connectors inspect source-google-pagespeed-insights --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Google PageSpeed Insights documentation: https://docs.airbyte.com/integrations/sources/google-pagespeed-insights

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
