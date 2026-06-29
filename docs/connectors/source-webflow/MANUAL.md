# pm connectors inspect source-webflow

```text
NAME
  pm connectors inspect source-webflow - Webflow connector manual

SYNOPSIS
  pm connectors inspect source-webflow
  pm connectors inspect source-webflow --json
  pm credentials add <name> --connector source-webflow [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Webflow catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/webflow.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.webflow.com/data/reference

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
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
  family: custom_go_port
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Webflow Data API: https://developers.webflow.com/data/reference
  Webflow authentication: https://developers.webflow.com/data/docs/getting-started
  Webflow API Changelog: https://developers.webflow.com/data/changelog
  Webflow v1 API Deprecation Notice: https://developers.webflow.com/data/docs/webflow-v1-api-deprecation-notice
  Webflow rate limits: https://developers.webflow.com/data/docs/rate-limits

CONFIGURATION
  accept_version (string): The version of the Webflow API to use. See https://developers.webflow.com/#versioning
  api_key (string) required secret: The API token for authenticating to Webflow. See https://university.webflow.com/lesson/intro-to-the-webflow-api
  site_id (string) required: The id of the Webflow site you are requesting data from. See https://developers.webflow.com/#sites
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
  pm connectors inspect source-webflow

  # Inspect as JSON
  pm connectors inspect source-webflow --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Webflow Data API: https://developers.webflow.com/data/reference
  Webflow authentication: https://developers.webflow.com/data/docs/getting-started
  Webflow API Changelog: https://developers.webflow.com/data/changelog
  Webflow v1 API Deprecation Notice: https://developers.webflow.com/data/docs/webflow-v1-api-deprecation-notice
  Webflow rate limits: https://developers.webflow.com/data/docs/rate-limits

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
