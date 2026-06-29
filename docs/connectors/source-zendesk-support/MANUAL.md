# pm connectors inspect source-zendesk-support

```text
NAME
  pm connectors inspect source-zendesk-support - Zendesk Support connector manual

SYNOPSIS
  pm connectors inspect source-zendesk-support
  pm connectors inspect source-zendesk-support --json
  pm credentials add <name> --connector source-zendesk-support [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Zendesk Support catalog connector. Native implementation status: planned_native_port.

ICON
  asset: icons/zendesk-support.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.zendesk.com/api-reference/ticketing/introduction/

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
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
  family: declarative_http_source
  priority_wave: 1
  etl_operations: catalog, check, read_snapshot
  reverse_etl_operations: none until native write conformance passes
  conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

OFFICIAL APPLICATION DOCUMENTATION
  Zendesk Support API: https://developer.zendesk.com/api-reference/ticketing/introduction/
  Zendesk authentication: https://developer.zendesk.com/api-reference/ticketing/introduction/#security-and-authentication
  API Changelog: https://developer.zendesk.com/api-reference/changelog/changelog/
  Zendesk API changelog: https://developer.zendesk.com/api-reference/ticketing/introduction/#changes
  Zendesk rate limits: https://developer.zendesk.com/api-reference/ticketing/account-configuration/usage_limits/
  Zendesk Status: https://status.zendesk.com/

CONFIGURATION
  credentials (object): manual intervention needed
  ignore_pagination (boolean): [Deprecated] Makes each stream read a single page of data.
  num_workers (integer): The number of worker threads to use for the sync. Higher values can improve sync throughput on large workspaces; lower values reduce load on the source.
  page_size (integer): The number of records per page for the ticket_comments stream API requests. Lower values may help prevent timeouts on large datasets. The maximum value is 1000.
  start_date (string): The UTC date and time from which you'd like to replicate data, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated.
  subdomain (string) required: This is your unique Zendesk subdomain that can be found in your account URL. For example, in https://MY_SUBDOMAIN.zendesk.com/, MY_SUBDOMAIN is the value of your subdomain.
  secret fields: credentials.access_token, credentials.api_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-zendesk-support

  # Inspect as JSON
  pm connectors inspect source-zendesk-support --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Zendesk Support API: https://developer.zendesk.com/api-reference/ticketing/introduction/
  Zendesk authentication: https://developer.zendesk.com/api-reference/ticketing/introduction/#security-and-authentication
  API Changelog: https://developer.zendesk.com/api-reference/changelog/changelog/
  Zendesk API changelog: https://developer.zendesk.com/api-reference/ticketing/introduction/#changes
  Zendesk rate limits: https://developer.zendesk.com/api-reference/ticketing/account-configuration/usage_limits/
  Zendesk Status: https://status.zendesk.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
