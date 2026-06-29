# pm connectors inspect source-sendgrid

```text
NAME
  pm connectors inspect source-sendgrid - Sendgrid connector manual

SYNOPSIS
  pm connectors inspect source-sendgrid
  pm connectors inspect source-sendgrid --json
  pm credentials add <name> --connector source-sendgrid [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Sendgrid catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/sendgrid.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/authentication

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
  API overview: https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/authentication
  SendGrid API OpenAPI specification: https://github.com/sendgrid/sendgrid-oai

CONFIGURATION
  api_key (string) required secret: Sendgrid API Key, use <a href=\"https://app.sendgrid.com/settings/api_keys/\">admin</a> to generate this key.
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
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
  pm connectors inspect source-sendgrid

  # Inspect as JSON
  pm connectors inspect source-sendgrid --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  API overview: https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/authentication
  SendGrid API OpenAPI specification: https://github.com/sendgrid/sendgrid-oai

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
