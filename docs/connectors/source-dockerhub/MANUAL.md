# pm connectors inspect source-dockerhub

```text
NAME
  pm connectors inspect source-dockerhub - Dockerhub connector manual

SYNOPSIS
  pm connectors inspect source-dockerhub
  pm connectors inspect source-dockerhub --json
  pm credentials add <name> --connector source-dockerhub [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Dockerhub catalog connector for https://docs.airbyte.com/integrations/sources/dockerhub. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
  support level: community

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-dockerhub:0.3.26 (metadata only; not executed)

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
  Docker Hub API reference: https://docs.docker.com/docker-hub/api/latest/
  Docker Hub authentication: https://docs.docker.com/docker-hub/api/latest/#section/Authentication
  Docker Hub rate limits: https://docs.docker.com/docker-hub/download-rate-limit/
  Docker Status: https://www.dockerstatus.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/dockerhub

CONFIGURATION
  docker_username (string) required: Username of DockerHub person or organization (for https://hub.docker.com/v2/repositories/USERNAME/ API call)

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/dockerhub

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-dockerhub

  # Inspect as JSON
  pm connectors inspect source-dockerhub --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Dockerhub documentation: https://docs.airbyte.com/integrations/sources/dockerhub

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
