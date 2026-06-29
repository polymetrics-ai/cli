# pm connectors inspect source-recurly

```text
NAME
  pm connectors inspect source-recurly - Recurly connector manual

SYNOPSIS
  pm connectors inspect source-recurly
  pm connectors inspect source-recurly --json
  pm credentials add <name> --connector source-recurly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Recurly catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/recurly.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.recurly.com/api/v2021-02-25/

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
  Recurly API reference: https://developers.recurly.com/api/v2021-02-25/
  Recurly authentication: https://developers.recurly.com/api/v2021-02-25/#section/Authentication
  Recurly rate limits: https://developers.recurly.com/api/v2021-02-25/#section/Rate-Limits

CONFIGURATION
  accounts_step_days (integer): Days in length for each API call to get data from the accounts stream. Smaller values will result in more API calls but better concurrency.
  api_key (string) required secret: manual intervention needed
  begin_time (string): ISO8601 timestamp from which the replication from Recurly API will start from.
  end_time (string): ISO8601 timestamp to which the replication from Recurly API will stop. Records after that date won't be imported.
  is_sandbox (boolean): Set to true for sandbox accounts (400 requests/min, all types). Defaults to false for production accounts (1,000 GET requests/min).
  num_workers (integer): The number of worker threads to use for the sync.
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
  pm connectors inspect source-recurly

  # Inspect as JSON
  pm connectors inspect source-recurly --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Recurly API reference: https://developers.recurly.com/api/v2021-02-25/
  Recurly authentication: https://developers.recurly.com/api/v2021-02-25/#section/Authentication
  Recurly rate limits: https://developers.recurly.com/api/v2021-02-25/#section/Rate-Limits

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
