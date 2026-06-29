# pm connectors inspect source-surveymonkey

```text
NAME
  pm connectors inspect source-surveymonkey - SurveyMonkey connector manual

SYNOPSIS
  pm connectors inspect source-surveymonkey
  pm connectors inspect source-surveymonkey --json
  pm credentials add <name> --connector source-surveymonkey [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  SurveyMonkey catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/surveymonkey.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.surveymonkey.com/api/v3/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  SurveyMonkey API reference: https://developer.surveymonkey.com/api/v3/
  SurveyMonkey authentication: https://developer.surveymonkey.com/api/v3/#authentication
  SurveyMonkey API Changelog: https://developer.surveymonkey.com/api/v3/#changelog
  SurveyMonkey rate limits: https://developer.surveymonkey.com/api/v3/#rate-limits

CONFIGURATION
  credentials (object) required: The authorization method to use to retrieve data from SurveyMonkey
  origin (string): Depending on the originating datacenter of the SurveyMonkey account, the API access URL may be different.
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  survey_ids (array): IDs of the surveys from which you'd like to replicate data. If left empty, data from all boards to which you have access will be replicated.
  secret fields: credentials.access_token, credentials.client_id, credentials.client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-surveymonkey

  # Inspect as JSON
  pm connectors inspect source-surveymonkey --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  SurveyMonkey API reference: https://developer.surveymonkey.com/api/v3/
  SurveyMonkey authentication: https://developer.surveymonkey.com/api/v3/#authentication
  SurveyMonkey API Changelog: https://developer.surveymonkey.com/api/v3/#changelog
  SurveyMonkey rate limits: https://developer.surveymonkey.com/api/v3/#rate-limits

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
