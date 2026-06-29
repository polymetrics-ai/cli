# pm connectors inspect source-kyve

```text
NAME
  pm connectors inspect source-kyve - KYVE connector manual

SYNOPSIS
  pm connectors inspect source-kyve
  pm connectors inspect source-kyve --json
  pm credentials add <name> --connector source-kyve [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  KYVE catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: alpha
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
  priority_wave: 3
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  KYVE documentation: https://docs.kyve.network/

CONFIGURATION
  max_pages (integer): The maximum amount of pages to go trough. Set to 'null' for all pages.
  page_size (integer): The pagesize for pagination, smaller numbers are used in integration tests.
  pool_ids (string) required: The IDs of the KYVE storage pool you want to archive. (Comma separated)
  start_ids (string) required: The start-id defines, from which bundle id the pipeline should start to extract the data. (Comma separated)
  url_base (string) required: URL to the KYVE Chain API.

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-kyve

  # Inspect as JSON
  pm connectors inspect source-kyve --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  KYVE documentation: https://docs.kyve.network/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
