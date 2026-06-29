# pm connectors inspect source-yandex-metrica

```text
NAME
  pm connectors inspect source-yandex-metrica - Yandex Metrica connector manual

SYNOPSIS
  pm connectors inspect source-yandex-metrica
  pm connectors inspect source-yandex-metrica --json
  pm credentials add <name> --connector source-yandex-metrica [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Yandex Metrica catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/yandexmetrica.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://yandex.com/dev/metrica/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: beta
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
  priority_wave: 2
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Yandex Metrica API: https://yandex.com/dev/metrica/

CONFIGURATION
  auth_token (string) required secret: Your Yandex Metrica API access token
  counter_id (string) required: Counter ID
  end_date (string): Starting point for your data replication, in format of "YYYY-MM-DD". If not provided will sync till most recent date.
  start_date (string) required: Starting point for your data replication, in format of "YYYY-MM-DD".
  secret fields: auth_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-yandex-metrica

  # Inspect as JSON
  pm connectors inspect source-yandex-metrica --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Yandex Metrica API: https://yandex.com/dev/metrica/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
