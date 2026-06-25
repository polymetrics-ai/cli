# pm connectors inspect source-elasticsearch

```text
NAME
  pm connectors inspect source-elasticsearch - Elasticsearch connector manual

SYNOPSIS
  pm connectors inspect source-elasticsearch
  pm connectors inspect source-elasticsearch --json
  pm credentials add <name> --connector source-elasticsearch [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Elasticsearch catalog connector for https://docs.airbyte.com/integrations/sources/elasticsearch. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: native_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-elasticsearch:0.1.5 (metadata only; not executed)

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
  Elasticsearch REST APIs: https://www.elastic.co/guide/en/elasticsearch/reference/current/rest-apis.html
  Elasticsearch authentication: https://www.elastic.co/guide/en/elasticsearch/reference/current/setting-up-authentication.html
  Elasticsearch Release Notes: https://www.elastic.co/docs/release-notes/elasticsearch
  Elastic Cloud Status: https://status.elastic.co/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/elasticsearch

CONFIGURATION
  authenticationMethod (object): The type of authentication to be used
  endpoint (string) required: The full url of the Elasticsearch server
  secret fields: authenticationMethod.apiKeySecret, authenticationMethod.password

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/elasticsearch

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-elasticsearch

  # Inspect as JSON
  pm connectors inspect source-elasticsearch --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Elasticsearch documentation: https://docs.airbyte.com/integrations/sources/elasticsearch

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
