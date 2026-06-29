# pm connectors inspect source-twilio

```text
NAME
  pm connectors inspect source-twilio - Twilio connector manual

SYNOPSIS
  pm connectors inspect source-twilio
  pm connectors inspect source-twilio --json
  pm credentials add <name> --connector source-twilio [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Twilio catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/twilio.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.twilio.com/docs/usage/api

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
  Twilio API reference: https://www.twilio.com/docs/usage/api
  Twilio authentication: https://www.twilio.com/docs/iam/api-keys
  Twilio Changelog: https://www.twilio.com/en-us/changelog
  Twilio rate limits: https://www.twilio.com/docs/usage/api#rate-limiting
  Twilio API OpenAPI specification: https://github.com/twilio/twilio-oai
  Twilio Status: https://status.twilio.com/

CONFIGURATION
  account_sid (string) required secret: Twilio account SID
  auth_token (string) required secret: Twilio Auth Token.
  lookback_window (integer): How far into the past to look for records. (in minutes)
  num_workers (integer): The number of worker threads to use for the sync.
  slice_step_duration (string): The time window size for each data slice when syncing incremental streams. Smaller windows may help avoid timeouts for accounts with large data volumes.
  start_date (string) required: UTC date and time in the format 2020-10-01T00:00:00Z. Any data before this date will not be replicated.
  secret fields: account_sid, auth_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-twilio

  # Inspect as JSON
  pm connectors inspect source-twilio --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Twilio API reference: https://www.twilio.com/docs/usage/api
  Twilio authentication: https://www.twilio.com/docs/iam/api-keys
  Twilio Changelog: https://www.twilio.com/en-us/changelog
  Twilio rate limits: https://www.twilio.com/docs/usage/api#rate-limiting
  Twilio API OpenAPI specification: https://github.com/twilio/twilio-oai
  Twilio Status: https://status.twilio.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
