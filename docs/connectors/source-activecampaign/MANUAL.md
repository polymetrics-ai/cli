# pm connectors inspect source-activecampaign

```text
NAME
  pm connectors inspect source-activecampaign - ActiveCampaign connector manual

SYNOPSIS
  pm connectors inspect source-activecampaign
  pm connectors inspect source-activecampaign --json
  pm credentials add <name> --connector source-activecampaign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  ActiveCampaign catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/activecampaign.svg
  source: official
  review_status: official_verified
  review_url: https://developers.activecampaign.com/reference/overview

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

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
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  ActiveCampaign documentation: https://developers.activecampaign.com/reference/overview

CONFIGURATION
  account_username (string) required: Account Username
  api_key (string) required secret: API Key
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
  pm connectors inspect source-activecampaign

  # Inspect as JSON
  pm connectors inspect source-activecampaign --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  ActiveCampaign documentation: https://developers.activecampaign.com/reference/overview

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
