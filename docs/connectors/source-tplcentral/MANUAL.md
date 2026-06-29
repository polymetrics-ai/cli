# pm connectors inspect source-tplcentral

```text
NAME
  pm connectors inspect source-tplcentral - TPLcentral connector manual

SYNOPSIS
  pm connectors inspect source-tplcentral
  pm connectors inspect source-tplcentral --json
  pm credentials add <name> --connector source-tplcentral [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  TPLcentral catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/tplcentral.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.3plcentral.com/

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
  TPL Central API: https://api.3plcentral.com/

CONFIGURATION
  client_id (string) required
  client_secret (string) required secret
  customer_id (integer)
  facility_id (integer)
  start_date (string): Date and time together in RFC 3339 format, for example, 2018-11-13T20:20:39+00:00.
  tpl_key (string)
  url_base (string) required
  user_login (string): User login ID and/or name is required
  user_login_id (integer): User login ID and/or name is required
  secret fields: client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-tplcentral

  # Inspect as JSON
  pm connectors inspect source-tplcentral --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  TPL Central API: https://api.3plcentral.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
