# pm connectors inspect source-slack

```text
NAME
  pm connectors inspect source-slack - Slack connector manual

SYNOPSIS
  pm connectors inspect source-slack
  pm connectors inspect source-slack --json
  pm credentials add <name> --connector source-slack [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Slack catalog connector for https://docs.airbyte.com/integrations/sources/slack. Native implementation status: planned_native_port.

CAPABILITIES
  catalog_metadata=true
  connector type: source
  release stage: generally_available
  support level: certified

IMPLEMENTATION STATUS
  implementation_status: planned_native_port
  runtime_kind: declarative_http_go
  notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
  upstream image reference: airbyte/source-slack:3.2.11 (metadata only; not executed)

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
  Changelog: https://api.slack.com/changelog
  Slack developer changelog: https://docs.slack.dev/changelog/
  Slack Web API OpenAPI specification: https://api.slack.com/specs/openapi/v2/slack_web.json
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/slack

CONFIGURATION
  channel_filter (array): A channel name list (without leading '#' char) which limit the channels from which you'd like to sync. Empty list means no filter.
  channel_messages_window_size (integer): The size (in days) of the date window that will be used while syncing data from the channel messages stream. A smaller window will allow for greater parallelization when syncing...
  credentials (object): Choose how to authenticate into Slack
  include_archived_channels (boolean): Whether to include archived channels in the sync. When disabled (default), archived channels are excluded from the Slack API response, reducing the number of API calls for downs...
  include_private_channels (boolean): Whether to read information from private channels that the bot is already in. If false, only public channels will be read. If true, the bot must be manually added to private cha...
  join_channels (boolean) required: Whether to join all channels or to sync data only from channels the bot is already in. If false, you''ll need to manually add the bot to all the channels from which you''d like ...
  lookback_window (integer) required: How far into the past to look for messages in threads, default is 0 days
  num_workers (integer): The number of worker threads to use for the sync.
  start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
  threads_ignore_no_replies (boolean): When enabled, the threads stream will skip messages that have no replies (reply_count is 0, null, or absent), reducing the number of API calls. Disabled by default to make Threa...
  secret fields: credentials.access_token, credentials.api_token, credentials.client_secret

SYNC MODES
  supported sync modes: full_refresh
  supports incremental: false

SECURITY
  Secret values are never rendered; only secret field names are shown.
  Upstream image references are metadata only and are not executed by pm.
  Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

DOCUMENTATION
  https://docs.airbyte.com/integrations/sources/slack

EXAMPLES
  # Inspect catalog entry
  pm connectors inspect source-slack

  # Inspect as JSON
  pm connectors inspect source-slack --json

AGENT WORKFLOW
  - Read implementation_status before planning ETL or reverse ETL.
  - If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
  - Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

SEE ALSO
  Slack documentation: https://docs.airbyte.com/integrations/sources/slack

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
