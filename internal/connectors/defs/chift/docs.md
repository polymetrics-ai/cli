# Overview

Reads and writes Chift consumers, connections, syncs, integrations, datastores, and webhook event
definitions through the Chift REST API using a session-token (client credentials) exchange.

Readable streams: `consumers`, `connections`, `syncs`, `integrations`, `datastores`,
`webhook_definitions`.

Write actions: `create_consumer`, `update_consumer`, `delete_consumer`.

Service API documentation: https://docs.chift.eu/api-reference.

## Auth setup

Connection fields:

- `account_id` (required, secret, string); Chift account ID sent in the session-token exchange
  request body. Never logged.
- `base_url` (optional, string); default `https://api.chift.eu`; format `uri`; Chift API base URL
  override for tests or proxies.
- `client_id` (required, secret, string); Chift OAuth client ID; used to obtain a session access
  token via POST /token. Never logged.
- `client_secret` (required, secret, string); Chift OAuth client secret; used to obtain a session
  access token via POST /token. Never logged.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).

Secret fields are redacted in logs and write previews: `account_id`, `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.chift.eu`, `page_size=100`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/consumers` with query `size`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `size`; page
size 100.

Pagination by stream: none: `integrations`, `datastores`, `webhook_definitions`; offset_limit:
`consumers`, `connections`, `syncs`.

- `consumers`: GET `/consumers` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `size`; page size 100.
- `connections`: GET `/connections` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `size`; page size 100.
- `syncs`: GET `/syncs` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `size`; page size 100.
- `integrations`: GET `/integrations` - records at response root.
- `datastores`: GET `/datastores` - records at response root.
- `webhook_definitions`: GET `/webhooks` - records at response root.

## Write actions & risks

Overall write risk: external mutation of Chift consumer records (create/update/delete); approval
required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_consumer`: POST `/consumers` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `email`, `internal_reference`, `name`, `redirect_url`; risk: external
  mutation; approval required.
- `update_consumer`: PATCH `/consumers/{{ record.consumerid }}` - kind `update`; body type `json`;
  path fields `consumerid`; required record fields `consumerid`; accepted fields `consumerid`,
  `email`, `internal_reference`, `name`, `redirect_url`; risk: external mutation; approval required.
- `delete_consumer`: DELETE `/consumers/{{ record.consumerid }}` - kind `delete`; body type `none`;
  path fields `consumerid`; required record fields `consumerid`; accepted fields `consumerid`;
  missing records treated as success for status `404`; risk: irreversible external deletion;
  approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, duplicate_of=4, non_data_endpoint=2, out_of_scope=24.
