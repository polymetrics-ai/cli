# pm connectors inspect source-paypal-transaction

```text
NAME
  pm connectors inspect source-paypal-transaction - Paypal Transaction connector manual

SYNOPSIS
  pm connectors inspect source-paypal-transaction
  pm connectors inspect source-paypal-transaction --json
  pm credentials add <name> --connector source-paypal-transaction [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Paypal Transaction catalog connector for https://docs.airbyte.com/integrations/sources/paypal-transaction. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-paypal-transaction:2.6.37 (metadata only; not executed)

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
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  PayPal API reference: https://developer.paypal.com/api/rest/
  PayPal authentication: https://developer.paypal.com/api/rest/authentication/
  PayPal rate limits: https://developer.paypal.com/api/rest/rate-limiting/
  PayPal Status: https://www.paypal-status.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/paypal-transaction

CONFIGURATION
  client_id (string) required secret: The Client ID of your Paypal developer application.
  client_secret (string) required secret: The Client Secret of your Paypal developer application.
  dispute_start_date (string): Start Date parameter for the list dispute endpoint in <a href=\"https://datatracker.ietf.org/doc/html/rfc3339#section-5.6\">ISO format</a>. This Start Date must be in range with...
  end_date (string): End Date for data extraction in <a href=\"https://datatracker.ietf.org/doc/html/rfc3339#section-5.6\">ISO format</a>. This can be help you select specific range of time, mainly ...
  is_sandbox (boolean) required: Determines whether to use the sandbox or production environment.
  refresh_token (string) secret: The key to refresh the expired access token.
  start_date (string) required: Start Date for data extraction in <a href=\"https://datatracker.ietf.org/doc/html/rfc3339#section-5.6\">ISO format</a>. Date must be in range from 3 years till 12 hrs before pre...
  time_window (integer): The number of days per request. Must be a number between 1 and 31.
  secret fields: client_id, client_secret, refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/paypal-transaction

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-paypal-transaction

  # Inspect as JSON
  pm connectors inspect source-paypal-transaction --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Paypal Transaction documentation: https://docs.airbyte.com/integrations/sources/paypal-transaction

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
