# pm connectors inspect source-the-guardian-api

```text
NAME
  pm connectors inspect source-the-guardian-api - The Guardian API connector manual

SYNOPSIS
  pm connectors inspect source-the-guardian-api
  pm connectors inspect source-the-guardian-api --json
  pm credentials add <name> --connector source-the-guardian-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  The Guardian API catalog connector for https://docs.airbyte.com/integrations/sources/the-guardian-api. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-the-guardian-api:0.2.26 (metadata only; not executed)

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
  The Guardian Open Platform: https://open-platform.theguardian.com/documentation/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/the-guardian-api

CONFIGURATION
  api_key (string) required secret: Your API Key. See <a href="https://open-platform.theguardian.com/access/">here</a>. The key is case sensitive.
  end_date (string): (Optional) Use this to set the maximum date (YYYY-MM-DD) of the results. Results newer than the end_date will not be shown. Default is set to the current date (today) for increm...
  query (string): (Optional) The query (q) parameter filters the results to only those that include that search term. The q parameter supports AND, OR and NOT operators.
  section (string): (Optional) Use this to filter the results by a particular section. See <a href="https://content.guardianapis.com/sections?api-key=test">here</a> for a list of all sections, and ...
  start_date (string) required: Use this to set the minimum date (YYYY-MM-DD) of the results. Results older than the start_date will not be shown.
  tag (string): (Optional) A tag is a piece of data that is used by The Guardian to categorise content. Use this parameter to filter results by showing only the ones matching the entered tag. S...
  secret fields: api_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/the-guardian-api

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-the-guardian-api

  # Inspect as JSON
  pm connectors inspect source-the-guardian-api --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  The Guardian API documentation: https://docs.airbyte.com/integrations/sources/the-guardian-api

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
