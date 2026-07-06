# Overview

Reads Countercyclical investments, valuations, research memos, teams, assumptions, and pipelines,
and creates investments, through the Countercyclical REST API.

Readable streams: `investments`, `valuations`, `memos`, `teams`, `assumptions`, `pipelines`.

Write actions: `create_investment`.

Service API documentation: https://docs.countercyclical.io/developers/endpoints.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Countercyclical API key, sent as the apiKey query parameter
  on every request. Never logged.
- `base_url` (optional, string); default `https://api.countercyclical.io/v1`; format `uri`;
  Countercyclical API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.countercyclical.io/v1`.

Authentication behavior:

- API key authentication in query parameter `apiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/investments`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `investments`: GET `/investments` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `valuations`: GET `/valuations` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `memos`: GET `/memos` - records path `.`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `teams`: GET `/teams` - records path `.`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `assumptions`: GET `/assumptions` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `pipelines`: GET `/pipelines` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

Overall write risk: external mutation: creates a new Investment record in the caller's workspace; no
update/delete actions are exposed.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_investment`: POST `/integrations/make/actions/investments` - kind `create`; body type
  `json`; required record fields `tickerSymbol`; accepted fields `tickerSymbol`; risk: creates a new
  Investment in the caller's Countercyclical workspace via the Make-integration action endpoint (the
  only documented general-purpose creation endpoint; the functionally-identical Zapier-integration
  endpoint is not separately exposed, see api_surface.json); external mutation, no approval
  required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=3, non_data_endpoint=3.
