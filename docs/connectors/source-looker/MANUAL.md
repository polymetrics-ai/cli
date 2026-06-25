# pm connectors inspect source-looker

```text
NAME
  pm connectors inspect source-looker - Looker connector manual

SYNOPSIS
  pm connectors inspect source-looker
  pm connectors inspect source-looker --json
  pm credentials add <name> --connector source-looker [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Looker catalog connector for https://docs.airbyte.com/integrations/sources/looker. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-looker:1.0.34 (metadata only; not executed)

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
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Looker API reference: https://cloud.google.com/looker/docs/reference/looker-api/latest
  Looker authentication: https://cloud.google.com/looker/docs/api-auth
  Looker rate limits: https://cloud.google.com/looker/docs/api-rate-limits
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/looker

CONFIGURATION
  client_id (string) required: The Client ID is first part of an API3 key that is specific to each Looker user. See the <a href="https://docs.airbyte.com/integrations/sources/looker">docs</a> for more informa...
  client_secret (string) required secret: The Client Secret is second part of an API3 key.
  domain (string) required: Domain for your Looker account, e.g. airbyte.cloud.looker.com,looker.[clientname].com,IP address
  run_look_ids (array): The IDs of any Looks to run
  secret fields: client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/looker

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-looker

  # Inspect as JSON
  pm connectors inspect source-looker --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Looker documentation: https://docs.airbyte.com/integrations/sources/looker

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
