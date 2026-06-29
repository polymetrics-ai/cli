# pm connectors inspect source-braintree

```text
NAME
  pm connectors inspect source-braintree - Braintree connector manual

SYNOPSIS
  pm connectors inspect source-braintree
  pm connectors inspect source-braintree --json
  pm credentials add <name> --connector source-braintree [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Braintree catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/braintree.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy

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
  Braintree API reference: https://developer.paypal.com/braintree/docs/reference/overview
  Braintree authentication: https://developer.paypal.com/braintree/docs/start/authentication
  Server SDK Deprecation Policy: https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy
  Braintree API rate limits: https://developer.paypal.com/braintree/docs/reference/general/rate-limiting
  Braintree Status: https://status.braintreepayments.com/

CONFIGURATION
  environment (string) required: Environment specifies where the data will come from.
  merchant_id (string) required: manual intervention needed
  private_key (string) required secret: manual intervention needed
  public_key (string) required: manual intervention needed
  start_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  secret fields: private_key

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-braintree

  # Inspect as JSON
  pm connectors inspect source-braintree --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Braintree API reference: https://developer.paypal.com/braintree/docs/reference/overview
  Braintree authentication: https://developer.paypal.com/braintree/docs/start/authentication
  Server SDK Deprecation Policy: https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy
  Braintree API rate limits: https://developer.paypal.com/braintree/docs/reference/general/rate-limiting
  Braintree Status: https://status.braintreepayments.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
