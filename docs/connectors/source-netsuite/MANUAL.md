# pm connectors inspect source-netsuite

```text
NAME
  pm connectors inspect source-netsuite - Netsuite connector manual

SYNOPSIS
  pm connectors inspect source-netsuite
  pm connectors inspect source-netsuite --json
  pm credentials add <name> --connector source-netsuite [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Netsuite catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/netsuite.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/chapter_1540391670.html

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
  NetSuite REST API: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/chapter_1540391670.html
  NetSuite authentication: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_4389727047.html

CONFIGURATION
  consumer_key (string) required secret: Consumer key associated with your integration
  consumer_secret (string) required secret: Consumer secret associated with your integration
  object_types (array): The API names of the Netsuite objects you want to sync. Setting this speeds up the connection setup process by limiting the number of schemas that need to be retrieved from Nets...
  realm (string) required secret: Netsuite realm e.g. 2344535, as for `production` or 2344535_SB1, as for the `sandbox`
  start_datetime (string) required: Starting point for your data replication, in format of "YYYY-MM-DDTHH:mm:ssZ"
  token_key (string) required secret: Access token key
  token_secret (string) required secret: Access token secret
  window_in_days (integer): The amount of days used to query the data with date chunks. Set smaller value, if you have lots of data.
  secret fields: consumer_key, consumer_secret, realm, token_key, token_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-netsuite

  # Inspect as JSON
  pm connectors inspect source-netsuite --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  NetSuite REST API: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/chapter_1540391670.html
  NetSuite authentication: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_4389727047.html

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
