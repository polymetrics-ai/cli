# pm connectors inspect source-marketo

```text
NAME
  pm connectors inspect source-marketo - Marketo connector manual

SYNOPSIS
  pm connectors inspect source-marketo
  pm connectors inspect source-marketo --json
  pm credentials add <name> --connector source-marketo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Marketo catalog connector for https://docs.airbyte.com/integrations/sources/marketo. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-marketo:2.0.1 (metadata only; not executed)

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
  Marketo REST API: https://developers.marketo.com/rest-api/
  Marketo authentication: https://developers.marketo.com/rest-api/authentication/
  Marketo rate limits: https://developers.marketo.com/rest-api/marketo-integration-best-practices/#api_limits
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/marketo

CONFIGURATION
  client_id (string) required secret: The Client ID of your Marketo developer application. See <a href="https://docs.airbyte.com/integrations/sources/marketo"> the docs </a> for info on how to obtain this.
  client_secret (string) required secret: The Client Secret of your Marketo developer application. See <a href="https://docs.airbyte.com/integrations/sources/marketo"> the docs </a> for info on how to obtain this.
  domain_url (string) required secret: Your Marketo Base URL. See <a href="https://docs.airbyte.com/integrations/sources/marketo"> the docs </a> for info on how to obtain this.
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  secret fields: client_id, client_secret, domain_url

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/marketo

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-marketo

  # Inspect as JSON
  pm connectors inspect source-marketo --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Marketo documentation: https://docs.airbyte.com/integrations/sources/marketo

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
