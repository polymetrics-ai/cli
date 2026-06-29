# pm connectors inspect source-facebook-pages

```text
NAME
  pm connectors inspect source-facebook-pages - Facebook Pages connector manual

SYNOPSIS
  pm connectors inspect source-facebook-pages
  pm connectors inspect source-facebook-pages --json
  pm credentials add <name> --connector source-facebook-pages [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Facebook Pages catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/facebook.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.facebook.com/docs/pages/

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
  Facebook Pages API reference: https://developers.facebook.com/docs/pages/
  Facebook authentication guide: https://developers.facebook.com/docs/facebook-login/guides/access-tokens/
  Facebook Graph API changelog: https://developers.facebook.com/docs/graph-api/changelog/
  Facebook Platform Status: https://developers.facebook.com/status/

CONFIGURATION
  access_token (string) required secret: Facebook Page Access Token
  page_id (string) required: Page ID
  page_size (integer): The number of records per page for paginated streams (post, post_insights). Decrease if encountering "reduce the amount of data" errors.
  secret fields: access_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-facebook-pages

  # Inspect as JSON
  pm connectors inspect source-facebook-pages --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Facebook Pages API reference: https://developers.facebook.com/docs/pages/
  Facebook authentication guide: https://developers.facebook.com/docs/facebook-login/guides/access-tokens/
  Facebook Graph API changelog: https://developers.facebook.com/docs/graph-api/changelog/
  Facebook Platform Status: https://developers.facebook.com/status/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
