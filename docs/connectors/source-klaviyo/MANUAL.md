# pm connectors inspect source-klaviyo

```text
NAME
  pm connectors inspect source-klaviyo - Klaviyo connector manual

SYNOPSIS
  pm connectors inspect source-klaviyo
  pm connectors inspect source-klaviyo --json
  pm credentials add <name> --connector source-klaviyo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Klaviyo catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/klaviyo.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.klaviyo.com/en/docs/api_versioning_and_deprecation_policy

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Versioning docs: https://developers.klaviyo.com/en/docs/api_versioning_and_deprecation_policy
  Changelog: https://developers.klaviyo.com/en/docs/changelog_
  Changelog: https://developers.klaviyo.com/en/docs/changelog
  Developer Group: https://community.klaviyo.com/groups/developer-group-64

CONFIGURATION
  api_key (string) required secret: manual intervention needed
  disable_fetching_predictive_analytics (boolean): Certain streams like the profiles stream can retrieve predictive analytics data from Klaviyo's API. However, at high volume, this can lead to service availability issues on the ...
  lookback_window (integer): The number of days to look back when syncing data in incremental mode. This helps capture any late-arriving data. Only applies to the events_detailed stream.
  metric_ids (string): OPTIONAL: Comma-separated list of specific metric IDs to use for flow_series_reports and campaign_values_reports streams. If left empty, the connector will automatically fetch r...
  num_workers (integer): The number of worker threads to use for the sync. The performance upper boundary is based on the limit of your Klaviyo plan. More info about the rate limit plan tiers can be fou...
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated. This field is optional - if not provided, all data will be replicated.
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-klaviyo

  # Inspect as JSON
  pm connectors inspect source-klaviyo --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Versioning docs: https://developers.klaviyo.com/en/docs/api_versioning_and_deprecation_policy
  Changelog: https://developers.klaviyo.com/en/docs/changelog_
  Changelog: https://developers.klaviyo.com/en/docs/changelog
  Developer Group: https://community.klaviyo.com/groups/developer-group-64

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
