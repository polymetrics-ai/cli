# pm connectors inspect source-airtable

```text
NAME
  pm connectors inspect source-airtable - Airtable connector manual

SYNOPSIS
  pm connectors inspect source-airtable
  pm connectors inspect source-airtable --json
  pm credentials add <name> --connector source-airtable [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Airtable catalog connector for https://docs.airbyte.com/integrations/sources/airtable. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-airtable:4.6.30 (metadata only; not executed)

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
  API Deprecation Guidelines: https://support.airtable.com/docs/airtable-api-deprecation-guidelines
  Changelog: https://airtable.com/developers/web/api/changelog
  Community blog: https://community.airtable.com/development-apis-11
  OAuth reference: https://airtable.com/developers/web/api/oauth-reference#authorization-request
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/airtable

CONFIGURATION
  add_base_id_to_stream_name (boolean): When enabled, includes the base ID in stream names to ensure uniqueness. Use this if you have cloned Airtable bases with duplicate table names. Note that enabling this will chan...
  credentials (object)
  num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may hit rate limits. Airtable limits to 5 requests per second per base.
  secret fields: credentials.access_token, credentials.api_key, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/airtable

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-airtable

  # Inspect as JSON
  pm connectors inspect source-airtable --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Airtable documentation: https://docs.airbyte.com/integrations/sources/airtable

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
