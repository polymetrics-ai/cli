# pm connectors inspect source-typeform

```text
NAME
  pm connectors inspect source-typeform - Typeform connector manual

SYNOPSIS
  pm connectors inspect source-typeform
  pm connectors inspect source-typeform --json
  pm credentials add <name> --connector source-typeform [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Typeform catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/typeform.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.typeform.com/developers/changelog/

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
  Changelog: https://www.typeform.com/developers/changelog/

CONFIGURATION
  credentials (object) required
  form_ids (array): When this parameter is set, the connector will replicate data only from the input forms. Otherwise, all forms in your Typeform account will be replicated. You can find form IDs ...
  start_date (string): The date from which you'd like to replicate data for Typeform API, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated.
  secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-typeform

  # Inspect as JSON
  pm connectors inspect source-typeform --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Changelog: https://www.typeform.com/developers/changelog/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
