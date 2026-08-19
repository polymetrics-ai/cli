# Overview

Reads Segment workspace, source, and destination metadata through the Segment Public API.

Readable streams: `workspace`, `sources`, `destinations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.segmentapis.com/tag/Getting-Started.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Segment Public API access token, sent as a Bearer token
  (Authorization: Bearer <api_token>). Never logged.
- `base_url` (optional, string); default `https://api.segmentapis.com`; format `uri`; Segment Public
  API base URL override for tests, proxies, or a region-specific endpoint (e.g.
  https://api.eu1.segmentapis.com).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.segmentapis.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/`.

## Streams notes

Default pagination for collection streams: page-number pagination; page parameter `page`; size
parameter `page_size`; starts at 1; page size 100. The singleton workspace stream does not paginate.

- `workspace`: GET `/` - the token-scoped singleton workspace from records path `data.workspace`;
  does not paginate; emits one passthrough record.
- `sources`: GET `/sources` - records path `sources`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; emits passthrough records.
- `destinations`: GET `/destinations` - records path `destinations`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100; emits passthrough
  records.

## Provider-shape migration

The legacy `workspaces` list stream at GET `/workspaces` was **REMOVED**: it is not present in the
provider's authoritative OpenAPI. Segment Public API tokens are scoped to one workspace, which the
documented singleton GET `/` returns as `data.workspace`. This is a user-visible connector behavior
change; use the `workspace` stream instead.

## Write actions & risks

This connector is read-only. Read behavior: external Segment Public API read of workspace, source,
and destination metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=29, duplicate_of=1, non_data_endpoint=18, out_of_scope=60,
  requires_elevated_scope=76.
