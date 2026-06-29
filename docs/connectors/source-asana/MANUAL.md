# pm connectors inspect source-asana

```text
NAME
  pm connectors inspect source-asana - Asana connector manual

SYNOPSIS
  pm connectors inspect source-asana
  pm connectors inspect source-asana --json
  pm credentials add <name> --connector source-asana [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Asana catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/asana.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.asana.com/reference/rest-api-reference

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  API reference: https://developers.asana.com/reference/rest-api-reference
  Authentication: https://developers.asana.com/docs/authentication
  Rate limits: https://developers.asana.com/docs/rate-limits
  Asana Status: https://status.asana.com/

CONFIGURATION
  credentials (object): Choose how to authenticate to Github
  num_workers (integer): The number of worker threads to use for the sync. The performance upper boundary is based on the limit of your Asana pricing plan. More info about the rate limit tiers can be fo...
  organization_export_ids (array): Globally unique identifiers for the organization exports
  test_mode (boolean): This flag is used for testing purposes for certain streams that return a lot of data. This flag is not meant to be enabled for prod.
  secret fields: credentials.client_id, credentials.client_secret, credentials.personal_access_token, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-asana

  # Inspect as JSON
  pm connectors inspect source-asana --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  API reference: https://developers.asana.com/reference/rest-api-reference
  Authentication: https://developers.asana.com/docs/authentication
  Rate limits: https://developers.asana.com/docs/rate-limits
  Asana Status: https://status.asana.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
