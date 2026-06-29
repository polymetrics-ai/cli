# pm connectors inspect source-gong

```text
NAME
  pm connectors inspect source-gong - Gong connector manual

SYNOPSIS
  pm connectors inspect source-gong
  pm connectors inspect source-gong --json
  pm credentials add <name> --connector source-gong [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Gong catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/gong.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://us-66463.app.gong.io/settings/api/documentation

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
  Gong API reference: https://us-66463.app.gong.io/settings/api/documentation
  Gong authentication: https://us-66463.app.gong.io/settings/api/documentation#overview
  Gong rate limits: https://us-66463.app.gong.io/settings/api/documentation#rate-limits

CONFIGURATION
  credentials (object) required: Choose how to authenticate to Gong.
  num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may increase API rate limit usage. Adjust based on your Gong API plan; the default of 4 is tuned t...
  start_date (string): The date from which to list calls, in the ISO-8601 format; if not specified, the calls start with the earliest recorded call. For web-conference calls recorded by Gong, the date...
  secret fields: credentials.access_key, credentials.access_key_secret, credentials.access_token, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-gong

  # Inspect as JSON
  pm connectors inspect source-gong --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Gong API reference: https://us-66463.app.gong.io/settings/api/documentation
  Gong authentication: https://us-66463.app.gong.io/settings/api/documentation#overview
  Gong rate limits: https://us-66463.app.gong.io/settings/api/documentation#rate-limits

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
